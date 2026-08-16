package service

import "testing"

func TestSessionRunning(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		session string
		want    bool
	}{
		{"matching session", "There is a screen on:\n\t12345.minecraft\t(Detached)\n1 Socket in /run/screen/S-user.\n", "minecraft", true},
		{"no sessions", "No Sockets found in /run/screen/S-user.\n", "minecraft", false},
		{"sibling prefix", "There is a screen on:\n\t12345.minecraft2\t(Detached)\n1 Socket in /run/screen/S-user.\n", "minecraft", false},
		{"sibling suffix", "There is a screen on:\n\t12345.backup-minecraft\t(Detached)\n1 Socket in /run/screen/S-user.\n", "minecraft", false},
		{"different session", "There is a screen on:\n\t12345.sodium\t(Detached)\n", "minecraft", false},
		{"detached flag", "There is a screen on:\n\t12345.minecraft\t(Detached)\n", "minecraft", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sessionRunning(tt.output, tt.session); got != tt.want {
				t.Errorf("sessionRunning(%q, %q) = %v, want %v", tt.output, tt.session, got, tt.want)
			}
		})
	}
}