package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"runtime"
	"strings"

	"github.com/eIGato/go-allure-report/parser"
)

// AllureTestSuites is a collection of Allure test suites.
type AllureTestSuites struct {
	XMLName xml.Name `xml:"testsuites"`
	Suites  []AllureTestSuite
}

// AllureTestSuite is a single Allure test suite which may contain many
// testcases.
type AllureTestSuite struct {
	XMLName    xml.Name        `xml:"testsuite"`
	Tests      int             `xml:"tests,attr"`
	Failures   int             `xml:"failures,attr"`
	Time       string          `xml:"time,attr"`
	Name       string          `xml:"name,attr"`
	Properties []AllureProperty `xml:"properties>property,omitempty"`
	TestCases  []AllureTestCase
}

// AllureTestCase is a single test case with its result.
type AllureTestCase struct {
	XMLName     xml.Name          `xml:"testcase"`
	Classname   string            `xml:"classname,attr"`
	Name        string            `xml:"name,attr"`
	Time        string            `xml:"time,attr"`
	SkipMessage *AllureSkipMessage `xml:"skipped,omitempty"`
	Failure     *AllureFailure     `xml:"failure,omitempty"`
}

// AllureSkipMessage contains the reason why a testcase was skipped.
type AllureSkipMessage struct {
	Message string `xml:"message,attr"`
}

// AllureProperty represents a key/value pair used to define properties.
type AllureProperty struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

// AllureFailure contains data related to a failed test.
type AllureFailure struct {
	Message  string `xml:"message,attr"`
	Type     string `xml:"type,attr"`
	Contents string `xml:",chardata"`
}

// AllureReportXML writes a Allure xml representation of the given report to w
// in the format described at http://windyroad.org/dl/Open%20Source/Allure.xsd
func AllureReportXML(report *parser.Report, noXMLHeader bool, goVersion string, w io.Writer) error {
	suites := AllureTestSuites{}

	// convert Report to Allure test suites
	for _, pkg := range report.Packages {
		ts := AllureTestSuite{
			Tests:      len(pkg.Tests),
			Failures:   0,
			Time:       formatTime(pkg.Time),
			Name:       pkg.Name,
			Properties: []AllureProperty{},
			TestCases:  []AllureTestCase{},
		}

		classname := pkg.Name
		if idx := strings.LastIndex(classname, "/"); idx > -1 && idx < len(pkg.Name) {
			classname = pkg.Name[idx+1:]
		}

		// properties
		if goVersion == "" {
			// if goVersion was not specified as a flag, fall back to version reported by runtime
			goVersion = runtime.Version()
		}
		ts.Properties = append(ts.Properties, AllureProperty{"go.version", goVersion})
		if pkg.CoveragePct != "" {
			ts.Properties = append(ts.Properties, AllureProperty{"coverage.statements.pct", pkg.CoveragePct})
		}

		// individual test cases
		for _, test := range pkg.Tests {
			testCase := AllureTestCase{
				Classname: classname,
				Name:      test.Name,
				Time:      formatTime(test.Time),
				Failure:   nil,
			}

			if test.Result == parser.FAIL {
				ts.Failures++
				testCase.Failure = &AllureFailure{
					Message:  "Failed",
					Type:     "",
					Contents: strings.Join(test.Output, "\n"),
				}
			}

			if test.Result == parser.SKIP {
				testCase.SkipMessage = &AllureSkipMessage{strings.Join(test.Output, "\n")}
			}

			ts.TestCases = append(ts.TestCases, testCase)
		}

		suites.Suites = append(suites.Suites, ts)
	}

	// to xml
	bytes, err := xml.MarshalIndent(suites, "", "\t")
	if err != nil {
		return err
	}

	writer := bufio.NewWriter(w)

	if !noXMLHeader {
		writer.WriteString(xml.Header)
	}

	writer.Write(bytes)
	writer.WriteByte('\n')
	writer.Flush()

	return nil
}

func formatTime(time int) string {
	return fmt.Sprintf("%.3f", float64(time)/1000.0)
}
