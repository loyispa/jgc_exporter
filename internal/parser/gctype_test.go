package parser

import "testing"

func TestParserGCTypeMethods(t *testing.T) {
	tests := []struct {
		name   string
		parser Parser
		want   GCType
	}{
		{"G1", NewG1Parser(false), GCTypeG1},
		{"G1-unified", NewG1Parser(true), GCTypeG1},
		{"CMS", NewCMSParser(false), GCTypeCMS},
		{"CMS-unified", NewCMSParser(true), GCTypeCMS},
		{"ZGC", NewZGCParser(), GCTypeZGC},
		{"Parallel", NewParallelParser(false), GCTypeParallel},
		{"Parallel-unified", NewParallelParser(true), GCTypeParallel},
		{"Serial", NewSerialParser(false), GCTypeSerial},
		{"Serial-unified", NewSerialParser(true), GCTypeSerial},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.parser.GCType(); got != tt.want {
				t.Errorf("GCType() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestFlushOnEmptyParsers(t *testing.T) {
	parsers := []Parser{
		NewG1Parser(false),
		NewG1Parser(true),
		NewCMSParser(false),
		NewCMSParser(true),
		NewZGCParser(),
		NewParallelParser(false),
		NewParallelParser(true),
		NewSerialParser(false),
		NewSerialParser(true),
	}
	for _, p := range parsers {
		p.SetHandler(&testHandler{})
		p.Flush()
	}
}

func TestSetHandlerAllParsers(t *testing.T) {
	h := &testHandler{}
	parsers := []Parser{
		NewG1Parser(false),
		NewCMSParser(false),
		NewZGCParser(),
		NewParallelParser(false),
		NewSerialParser(false),
	}
	for _, p := range parsers {
		p.SetHandler(h)
	}
}
