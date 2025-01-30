package vcsretriever

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeepseekFilter(t *testing.T) {
	tests := []struct {
		name string
		args string
		want string
	}{
		{
			name: "working case",
			args: `
<think>
toto
</think>
tutu`,
			want: "\ntutu",
		},
		{
			name: "working case with return line",
			args: `
<think>
toto
</think>

tutu`,
			want: "\n\ntutu",
		},
		{
			name: "working case inline",
			args: `
<think>
toto
</think>tutu
tutu`,
			want: `
tutu
tutu`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := DeepseekFilter()
			assert.NoError(t, err)
			got := filter(tt.args)
			assert.Equal(t, tt.want, got)
		})
	}
}
