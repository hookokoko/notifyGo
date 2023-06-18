package clientX

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestService_GetTargets(t *testing.T) {
	tests := []struct {
		name          string
		ServiceConfig *ServiceConfig
		want          []*Addr
		wantErr       assert.ErrorAssertionFunc
	}{
		{
			name: "base",
			ServiceConfig: func() *ServiceConfig {
				config, _ := loadService("conf/test.toml")
				return config
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				IDC:           "test",
				ServiceConfig: tt.ServiceConfig,
			}
			got, err := s.GetTargets()
			assert.Nil(t, err)
			t.Log(got)
		})
	}
}

func TestService_PickTarget(t *testing.T) {
	tests := []struct {
		name          string
		IDC           string
		ServiceConfig *ServiceConfig
	}{
		{
			name: "base",
			IDC:  "test",
			ServiceConfig: func() *ServiceConfig {
				config, _ := loadService("conf/test.toml")
				return config
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				IDC:           tt.IDC,
				ServiceConfig: tt.ServiceConfig,
			}
			targets, err := s.GetTargets()
			assert.Nil(t, err)
			s.Bala = NewBalanceBuilder[*Addr](s.Name, targets).Build("roundrobin")
			for i := 0; i < 10; i++ {
				target := s.PickTarget()
				t.Log(target)
			}

		})
	}
}
