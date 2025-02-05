package runner

type StarlarktestError = []interface{}

type TestSuiteSummary struct {
	Project *KurtestosisProject
	summaries []TestFileSummary
}

func NewTestSuiteSummary(project *KurtestosisProject) *TestSuiteSummary {
	return &TestSuiteSummary{
		Project: project,
	}
}

func (summary *TestSuiteSummary) Append(testFileSummary *TestFileSummary) {
	summary.summaries = append(summary.summaries, *testFileSummary)
}

func (summary *TestSuiteSummary) Summaries() []TestFileSummary {
	return summary.summaries
}

func (summary *TestSuiteSummary) Success() bool {
	for _, testFileSummary := range(summary.summaries) {
		if !testFileSummary.Success() {
			return false
		}
	}

	return true
}

type TestFileSummary struct {
	TestFile *TestFile
	summaries []TestFunctionSummary
}

func (summary *TestFileSummary) Summaries() []TestFunctionSummary {
	return summary.summaries
}

func (summary *TestFileSummary) Append(testFunctionSummary *TestFunctionSummary) {
	summary.summaries = append(summary.summaries, *testFunctionSummary)
}

func (summary *TestFileSummary) Success() bool {
	for _, testFunctionSummary := range(summary.summaries) {
		if !testFunctionSummary.Success() {
			return false
		}
	}

	return true
}

type TestFunctionSummary struct {
	TestFunction *TestFunction
	errors []StarlarktestError
}

func (summary *TestFunctionSummary) Errors() []StarlarktestError {
	return summary.errors
}

func (summary *TestFunctionSummary) Success() bool {
	return len(summary.errors) == 0
}

type StarlarktestReporter struct {
	TestFunction *TestFunction
	errors []StarlarktestError
}

func (reporter *StarlarktestReporter) Error(args ...interface{}) {
	reporter.errors = append(reporter.errors, args)
}

func (reporter *StarlarktestReporter) Summary() *TestFunctionSummary {
	return &TestFunctionSummary{
		TestFunction: reporter.TestFunction,
		errors: reporter.errors,
	}
}

func NewStarlarktestReporter(testFunction *TestFunction) *StarlarktestReporter {
	return &StarlarktestReporter{
		TestFunction: testFunction,
	}
}

func NewTestFileSummary(testFile *TestFile) *TestFileSummary {
	return &TestFileSummary{
		TestFile: testFile,
	}
}