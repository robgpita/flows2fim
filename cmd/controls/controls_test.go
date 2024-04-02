package controls

import (
	"testing"

	_ "modernc.org/sqlite"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"positive", []string{"-db=/app/testdata/reach_data.db", "-f", "/app/testdata/flows_100yr.csv", "-c", "/app/testdata/outputs/controls.csv", "-sid", "8489318", "-scs", "0.0"}, false},
		{"db file does not exist", []string{"-db=/app/testdata/not_exist.db", "-f", "/app/testdata/flows_100yr.csv", "-c", "/app/testdata/outputs/controls.csv", "-sid", "8489318", "-scs", "0.0"}, true},
		{"flow file does not exist", []string{"-db=/app/testdata/reach_data.db", "-f", "/app/testdata/flows_not_exist.csv", "-c", "/app/testdata/outputs/controls.csv", "-sid", "8489318", "-scs", "0.0"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Run(tt.args); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
