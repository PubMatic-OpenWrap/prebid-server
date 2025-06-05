package sdkparser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseTemplateAndSetValues(t *testing.T) {
	type args struct {
		template map[string]any
		source   map[string]any
		target   map[string]any
	}
	tests := []struct {
		name     string
		args     args
		expected map[string]any
	}{
		{
			name: "Basic fields",
			args: args{
				template: map[string]any{
					"id": "string",
					"device": map[string]any{
						"ip": "string",
					},
				},
				source: map[string]any{
					"id": "source_id",
				},
				target: map[string]any{},
			},
			expected: map[string]any{
				"id": "source_id",
			},
		},
		{
			name: "nested fields",
			args: args{
				template: map[string]any{
					"id": "string",
					"device": map[string]any{
						"ip": "string",
						"geo": map[string]any{
							"country": "string",
							"region":  "string",
							"city":    "string",
							"coordinates": map[string]any{
								"lat": "float",
								"lon": "float",
							},
						},
					},
				},
				source: map[string]any{
					"id": "source_id",
					"device": map[string]any{
						"ip": "127.0.0.1",
						"geo": map[string]any{
							"country": "US",
							"coordinates": map[string]any{
								"lat": 37.7749,
								"lon": -122.4194,
							},
						},
					},
				},
				target: map[string]any{
					"id": "target_id",
					"app": map[string]any{
						"name": "Test App",
					},
				},
			},
			expected: map[string]any{
				"id": "source_id",
				"app": map[string]any{
					"name": "Test App",
				},
				"device": map[string]any{
					"ip": "127.0.0.1",
					"geo": map[string]any{
						"country": "US",
						"coordinates": map[string]any{
							"lat": 37.7749,
							"lon": -122.4194,
						},
					},
				},
			},
		},
		{
			name: "nested fields update existing",
			args: args{
				template: map[string]any{
					"id": "string",
					"app": map[string]any{
						"name": "string",
					},
					"device": map[string]any{
						"ip": "string",
						"geo": map[string]any{
							"country": "string",
							"region":  "string",
							"city":    "string",
							"coordinates": map[string]any{
								"lat": "float",
								"lon": "float",
							},
						},
					},
				},
				source: map[string]any{
					"id": "source_id",
					"app": map[string]any{
						"name": "source_app",
					},
					"device": map[string]any{
						"ip": "127.0.0.1",
						"geo": map[string]any{
							"country": "US",
							"coordinates": map[string]any{
								"lat": 37.7749,
								"lon": -122.4194,
							},
						},
					},
				},
				target: map[string]any{
					"id": "target_id",
					"app": map[string]any{
						"name": "Target App",
					},
				},
			},
			expected: map[string]any{
				"id": "source_id",
				"app": map[string]any{
					"name": "source_app",
				},
				"device": map[string]any{
					"ip": "127.0.0.1",
					"geo": map[string]any{
						"country": "US",
						"coordinates": map[string]any{
							"lat": 37.7749,
							"lon": -122.4194,
						},
					},
				},
			},
		},
		{
			name: "nested fields update existing",
			args: args{
				template: map[string]any{
					"id": "string",
					"app": map[string]any{
						"name": "string",
					},
					"device": map[string]any{
						"ip": "string",
						"geo": map[string]any{
							"country": "string",
							"region":  "string",
							"city":    "string",
							"coordinates": map[string]any{
								"lat": "float",
								"lon": "float",
							},
						},
					},
				},
				source: map[string]any{
					"id": "source_id",
					"app": map[string]any{
						"name": "source_app",
					},
					"device": map[string]any{
						"ip": "127.0.0.1",
						"geo": map[string]any{
							"country": "US",
							"coordinates": map[string]any{
								"lat": 37.7749,
								"lon": -122.4194,
							},
						},
					},
				},
				target: map[string]any{
					"id": "target_id",
					"app": map[string]any{
						"name": "Target App",
					},
				},
			},
			expected: map[string]any{
				"id": "source_id",
				"app": map[string]any{
					"name": "source_app",
				},
				"device": map[string]any{
					"ip": "127.0.0.1",
					"geo": map[string]any{
						"country": "US",
						"coordinates": map[string]any{
							"lat": 37.7749,
							"lon": -122.4194,
						},
					},
				},
			},
		},
		{
			name: "Nested fields with arrays",
			args: args{
				template: map[string]any{
					"id": "string",
					"app": map[string]any{
						"name": "string",
					},
					"imp": []any{
						map[string]any{
							"id": "string",
							"banner": map[string]any{
								"w": "int",
								"h": "int",
							},
							"format": []any{"int"},
						},
					},
				},
				source: map[string]any{
					"id": "source_id",
					"app": map[string]any{
						"name": "source_app",
					},
					"imp": []any{
						map[string]any{
							"id": "imp1",
							"banner": map[string]any{
								"w": 250,
								"h": 350,
							},
							"format": []any{300, 200},
						},
					},
				},
				target: map[string]any{
					"id": "target_id",
					"app": map[string]any{
						"name": "Target App",
					},
					"imp": []any{
						map[string]any{
							"id": "imp1",
							"banner": map[string]any{
								"w": 450,
							},
						},
					},
				},
			},
			expected: map[string]any{
				"id": "source_id",
				"app": map[string]any{
					"name": "source_app",
				},
				"imp": []any{
					map[string]any{
						"id": "imp1",
						"banner": map[string]any{
							"w": 250,
							"h": 350,
						},
						"format": []any{300, 200},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ParseTemplateAndSetValues(tt.args.template, tt.args.source, tt.args.target)
			assert.Equal(t, tt.expected, tt.args.target)
		})
	}
}
