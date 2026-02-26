package main

import "testing"

func TestValidateConsoleCommand(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "ok", input: "say hello", wantErr: false},
		{name: "empty", input: "   ", wantErr: true},
		{name: "newline", input: "say a\nsay b", wantErr: true},
		{name: "carriage", input: "say a\rsay b", wantErr: true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := validateConsoleCommand(tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateConsoleCommand(%q) err=%v wantErr=%v", tc.input, err, tc.wantErr)
			}
		})
	}
}

func TestValidatePlayerName(t *testing.T) {
	t.Parallel()
	if err := validatePlayerName("Steve_123"); err != nil {
		t.Fatalf("expected valid player name, got error: %v", err)
	}
	if err := validatePlayerName("bad name"); err == nil {
		t.Fatal("expected invalid player name")
	}
}

func TestValidateFunctionID(t *testing.T) {
	t.Parallel()
	if err := validateFunctionID("mcserver:mainhall/hub_layout"); err != nil {
		t.Fatalf("expected valid function id, got error: %v", err)
	}
	if err := validateFunctionID("mcserver:bad function"); err == nil {
		t.Fatal("expected invalid function id")
	}
}
