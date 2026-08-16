package empirical

import _ "embed"

// ChartDBData is the embedded chart database (charts/people.json).
//
//go:embed charts/people.json
var ChartDBData []byte
