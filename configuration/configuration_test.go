package configuration

import (
	"os"
	"path"
	"reflect"
	"testing"
)

func TestGetConfiguration(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv(configEnv, tmpDir)

	tests := []struct {
		name        string
		fileContent string
		want        Configuration
		wantErr     bool
	}{
		{
			name:        "empty file",
			fileContent: "",
			want:        Configuration{},
			wantErr:     true,
		},
		{
			name: "empty telegram token",
			fileContent: `
telegram:
    token: ""
    passwd: "passwd"
`,
			want:    Configuration{},
			wantErr: true,
		},
		{
			name: "ok without http",
			fileContent: `
telegram:
    token: "token"
    passwd: "passwd"
radarr:
    apiKey: "apiKeyR"
    endpoint: "endpointR"
sonarr:
    apiKey: "apiKeyS"
    endpoint: "endpointS"
`,
			want: Configuration{
				Telegram: Telegram{
					Token:  "token",
					Passwd: "passwd",
				},
				Radarr: Radarr{
					ApiKey:   "apiKeyR",
					Endpoint: "http://endpointR",
				},
				Sonarr: Sonarr{
					ApiKey:  "apiKeyS",
					Endpoint: "http://endpointS",
				},
			},
			wantErr: false,
		},
		{
			name: "ok with http",
			fileContent: `
telegram:
    token: "token"
    passwd: "passwd"
radarr:
    apiKey: "apiKeyR"
    endpoint: "https://endpointR"
sonarr:
    apiKey: "apiKeyS"
    endpoint: "https://endpointS"
`,
			want: Configuration{
				Telegram: Telegram{
					Token:  "token",
					Passwd: "passwd",
				},
				Radarr: Radarr{
					ApiKey:   "apiKeyR",
					Endpoint: "https://endpointR",
				},
				Sonarr: Sonarr{
					ApiKey:  "apiKeyS",
					Endpoint: "https://endpointS",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// write the file
			filePath := path.Join(getConfigPath(), configFileName)
			err := os.WriteFile(filePath, []byte(tt.fileContent), 0644)
			if err != nil {
				t.Fatalf("error writing file: %v", err)
			}

			got, err := GetConfiguration()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConfiguration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetConfiguration() = %v, want %v", got, tt.want)
			}
		})
	}
}
