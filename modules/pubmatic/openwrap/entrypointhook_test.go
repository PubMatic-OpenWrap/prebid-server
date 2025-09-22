package openwrap

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/prebid/prebid-server/v3/hooks/hookexecution"
	"github.com/prebid/prebid-server/v3/hooks/hookstage"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/config"
	mock_metrics "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/metrics/mock"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/nbr"
	mock_feature "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/publisherfeature/mock"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/wakanda"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestOpenWrap_handleEntrypointHook(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
	mockFeature := mock_feature.NewMockFeature(ctrl)
	reset := wakanda.TestInstance("111", "222")

	originalOw := ow
	defer func() { ow = originalOw }()
	ow = &OpenWrap{pubFeatures: mockFeature, metricEngine: mockEngine}
	defer func() {
		ctrl.Finish()
		reset()
	}()

	type fields struct {
		cfg   config.Config
		cache cache.Cache
	}
	type args struct {
		in0     context.Context
		miCtx   hookstage.ModuleInvocationContext
		payload hookstage.EntrypointPayload
		setup   func(*mock_metrics.MockMetricsEngine)
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		want     hookstage.HookResult[hookstage.EntrypointPayload]
		wantBody []byte
		wantErr  error
		doMutate bool
	}{
		{
			name: "request with sshb=1 should not execute entrypointhook",
			args: args{
				in0:   context.Background(),
				miCtx: hookstage.ModuleInvocationContext{},
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						r, err := http.NewRequest("POST", "http://localhost/openrtb/2.5?sshb=1", nil)
						if err != nil {
							panic(err)
						}
						r.Header.Add("User-Agent", "go-test")
						r.Header.Add("SOURCE_IP", "127.0.0.1")
						r.Header.Add("Cookie", `KADUSERCOOKIE=7D75D25F-FAC9-443D-B2D1-B17FEE11E027; DPSync3=1684886400%3A248%7C1685491200%3A245_226_201; KRTBCOOKIE_80=16514-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&22987-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&23025-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&23386-CAESEMih0bN7ISRdZT8xX8LXzEw; KRTBCOOKIE_377=6810-59dc50c9-d658-44ce-b442-5a1f344d97c0&KRTB&22918-59dc50c9-d658-44ce-b442-5a1f344d97c0&KRTB&23031-59dc50c9-d658-44ce-b442-5a1f344d97c0; uids=eyJ0ZW1wVUlEcyI6eyIzM2Fjcm9zcyI6eyJ1aWQiOiIxMTkxNzkxMDk5Nzc2NjEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTo0My4zODg4Nzk5NVoifSwiYWRmIjp7InVpZCI6IjgwNDQ2MDgzMzM3Nzg4MzkwNzgiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMS4wMzMwNTQ3MjdaIn0sImFka2VybmVsIjp7InVpZCI6IkE5MTYzNTAwNzE0OTkyOTMyOTkwIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuMzczMzg1NjYyWiJ9LCJhZGtlcm5lbEFkbiI6eyJ1aWQiOiJBOTE2MzUwMDcxNDk5MjkzMjk5MCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEzLjQzNDkyNTg5NloifSwiYWRtaXhlciI6eyJ1aWQiOiIzNjZhMTdiMTJmMjI0ZDMwOGYzZTNiOGRhOGMzYzhhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjU5MjkxNDgwMVoifSwiYWRueHMiOnsidWlkIjoiNDE5Mjg5ODUzMDE0NTExOTMiLCJleHBpcmVzIjoiMjAyMy0wMS0xOFQwOTo1MzowOC44MjU0NDI2NzZaIn0sImFqYSI6eyJ1aWQiOiJzMnN1aWQ2RGVmMFl0bjJveGQ1aG9zS1AxVmV3IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTMuMjM5MTc2MDU0WiJ9LCJlcGxhbm5pbmciOnsidWlkIjoiQUoxRjBTOE5qdTdTQ0xWOSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjkyOTk2MDQ3M1oifSwiZ2Ftb3NoaSI6eyJ1aWQiOiJndXNyXzM1NmFmOWIxZDhjNjQyYjQ4MmNiYWQyYjdhMjg4MTYxIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuNTI0MTU3MjI1WiJ9LCJncmlkIjp7InVpZCI6IjRmYzM2MjUwLWQ4NTItNDU5Yy04NzcyLTczNTZkZTE3YWI5NyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE0LjY5NjMxNjIyN1oifSwiZ3JvdXBtIjp7InVpZCI6IjdENzVEMjVGLUZBQzktNDQzRC1CMkQxLUIxN0ZFRTExRTAyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjM5LjIyNjIxMjUzMloifSwiaXgiOnsidWlkIjoiWW9ORlNENlc5QkphOEh6eEdtcXlCUUFBXHUwMDI2Mjk3IiwiZXhwaXJlcyI6IjIwMjMtMDUtMzFUMDc6NTM6MzguNTU1ODI3MzU0WiJ9LCJqaXhpZSI6eyJ1aWQiOiI3MzY3MTI1MC1lODgyLTExZWMtYjUzOC0xM2FjYjdhZjBkZTQiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi4xOTEwOTk3MzJaIn0sImxvZ2ljYWQiOnsidWlkIjoiQVZ4OVROQS11c25pa3M4QURzTHpWa3JvaDg4QUFBR0JUREh0UUEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS40NTUxNDk2MTZaIn0sIm1lZGlhbmV0Ijp7InVpZCI6IjI5Nzg0MjM0OTI4OTU0MTAwMDBWMTAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMy42NzIyMTUxMjhaIn0sIm1naWQiOnsidWlkIjoibTU5Z1hyN0xlX1htIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTcuMDk3MDAxNDcxWiJ9LCJuYW5vaW50ZXJhY3RpdmUiOnsidWlkIjoiNmFlYzhjMTAzNzlkY2I3ODQxMmJjODBiNmRkOWM5NzMxNzNhYjdkNzEyZTQzMWE1YTVlYTcwMzRlNTZhNThhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE2LjcxNDgwNzUwNVoifSwib25ldGFnIjp7InVpZCI6IjdPelZoVzFOeC1LOGFVak1HMG52NXVNYm5YNEFHUXZQbnVHcHFrZ3k0ckEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS4xNDE3NDEyNjJaIn0sIm9wZW54Ijp7InVpZCI6IjVkZWNlNjIyLTBhMjMtMGRhYi0zYTI0LTVhNzcwMTBlNDU4MiIsImV4cGlyZXMiOiIyMDIzLTA1LTMxVDA3OjUyOjQ3LjE0MDQxNzM2M1oifSwicHVibWF0aWMiOnsidWlkIjoiN0Q3NUQyNUYtRkFDOS00NDNELUIyRDEtQjE3RkVFMTFFMDI3IiwiZXhwaXJlcyI6IjIwMjItMTAtMzFUMDk6MTQ6MjUuNzM3MjU2ODk5WiJ9LCJyaWNoYXVkaWVuY2UiOnsidWlkIjoiY2I2YzYzMjAtMzNlMi00Nzc0LWIxNjAtMXp6MTY1NDg0MDc0OSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjUyNTA3NDE4WiJ9LCJzbWFydHlhZHMiOnsidWlkIjoiMTJhZjE1ZTQ0ZjAwZDA3NjMwZTc0YzQ5MTU0Y2JmYmE0Zjg0N2U4ZDRhMTU0YzhjM2Q1MWY1OGNmNzJhNDYyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjgyNTAzMTg4NFoifSwic21pbGV3YW50ZWQiOnsidWlkIjoiZGQ5YzNmZTE4N2VmOWIwOWNhYTViNzExNDA0YzI4MzAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNC4yNTU2MDkzNjNaIn0sInN5bmFjb3JtZWRpYSI6eyJ1aWQiOiJHRFBSIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuOTc5NTgzNDM4WiJ9LCJ0cmlwbGVsaWZ0Ijp7InVpZCI6IjcwMjE5NzUwNTQ4MDg4NjUxOTQ2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA4Ljk4OTY3MzU3NFoifSwidmFsdWVpbXByZXNzaW9uIjp7InVpZCI6IjlkMDgxNTVmLWQ5ZmUtNGI1OC04OThlLWUyYzU2MjgyYWIzZSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjA2NzgzOTE2NFoifSwidmlzeCI6eyJ1aWQiOiIyN2UwYWMzYy1iNDZlLTQxYjMtOTkyYy1mOGQyNzE0OTQ5NWUiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi45ODk1MjM1NzNaIn0sInlpZWxkbGFiIjp7InVpZCI6IjY5NzE0ZDlmLWZiMDAtNGE1Zi04MTljLTRiZTE5MTM2YTMyNSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjExLjMwMzAyNjYxNVoifSwieWllbGRtbyI6eyJ1aWQiOiJnOTZjMmY3MTlmMTU1MWIzMWY2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjExMDUyODYwOVoifSwieWllbGRvbmUiOnsidWlkIjoiMmE0MmZiZDMtMmM3MC00ZWI5LWIxYmQtMDQ2OTY2NTBkOTQ4IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuMzE4MzMzOTM5WiJ9LCJ6ZXJvY2xpY2tmcmF1ZCI6eyJ1aWQiOiJiOTk5NThmZS0yYTg3LTJkYTQtOWNjNC05NjFmZDExM2JlY2UiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNS43MTk1OTQ1NjZaIn19LCJiZGF5IjoiMjAyMi0wNS0xN1QwNjo0ODozOC4wMTc5ODgyMDZaIn0=; KRTBCOOKIE_153=1923-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&19420-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&22979-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&23462-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse; KRTBCOOKIE_57=22776-41928985301451193&KRTB&23339-41928985301451193; KRTBCOOKIE_27=16735-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&16736-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&23019-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&23114-uid:3cab6283-4546-4500-a7b6-40ef605fe745; KRTBCOOKIE_18=22947-1978557989514665832; KRTBCOOKIE_466=16530-4fc36250-d852-459c-8772-7356de17ab97; KRTBCOOKIE_391=22924-8044608333778839078&KRTB&23263-8044608333778839078&KRTB&23481-8044608333778839078; KRTBCOOKIE_1310=23431-b81c3g7dr67i&KRTB&23446-b81c3g7dr67i&KRTB&23465-b81c3g7dr67i; KRTBCOOKIE_1290=23368-vkf3yv9lbbl; KRTBCOOKIE_22=14911-4554572065121110164&KRTB&23150-4554572065121110164; KRTBCOOKIE_860=16335-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23334-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23417-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23426-YGAqDU1zUTdjyAFxCoe3kctlNPo; KRTBCOOKIE_904=16787-KwJwE7NkCZClNJRysN2iYg; KRTBCOOKIE_1159=23138-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23328-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23427-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23445-5545f53f3d6e4ec199d8ed627ff026f3; KRTBCOOKIE_32=11175-AQEI_1QecY2ESAIjEW6KAQEBAQE&KRTB&22713-AQEI_1QecY2ESAIjEW6KAQEBAQE&KRTB&22715-AQEI_1QecY2ESAIjEW6KAQEBAQE; SyncRTB3=1685577600%3A35%7C1685491200%3A107_21_71_56_204_247_165_231_233_179_22_209_54_254_238_96_99_220_7_214_13_3_8_234_176_46_5%7C1684886400%3A2_223_15%7C1689465600%3A69%7C1685145600%3A63; KRTBCOOKIE_107=1471-uid:EK38R0PM1NQR0H5&KRTB&23421-uid:EK38R0PM1NQR0H5; KRTBCOOKIE_594=17105-RX-447a6332-530e-456a-97f4-3f0fd1ed48c9-004&KRTB&17107-RX-447a6332-530e-456a-97f4-3f0fd1ed48c9-004; SPugT=1684310122; chkChromeAb67Sec=133; KRTBCOOKIE_699=22727-AAFy2k7FBosAAEasbJoXnw; PugT=1684310473; origin=go-test`)
						return r
					}(),
					Body: []byte(`{"ext":{"wrapper":{"profileid":5890,"versionid":1}}}`),
				},
				setup: func(mme *mock_metrics.MockMetricsEngine) {},
			},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{
				ModuleContext: hookstage.ModuleContext{
					"rctx": models.RequestCtx{
						Sshb:      "1",
						DeviceCtx: models.DeviceCtx{IP: "127.0.0.1", UA: "go-test"},
					},
				},
			},
		},
		{
			name: "valid /openrtb2/2.5 request(reverseProxy through SSHB(8001->8000))",
			fields: fields{
				cfg: config.Config{
					Tracker: config.Tracker{
						Endpoint:                  "t.pubmatic.com",
						VideoErrorTrackerEndpoint: "t.pubmatic.com/error",
					},
				},
				cache: nil,
			},
			args: args{
				in0:   context.Background(),
				miCtx: hookstage.ModuleInvocationContext{},
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						r, err := http.NewRequest("POST", "http://localhost/openrtb/2.5?debug=1&sshb=2", nil)
						if err != nil {
							panic(err)
						}
						r.Header.Add("User-Agent", "go-test")
						r.Header.Add("SOURCE_IP", "127.0.0.1")
						r.Header.Add("Cookie", `KADUSERCOOKIE=7D75D25F-FAC9-443D-B2D1-B17FEE11E027; DPSync3=1684886400%3A248%7C1685491200%3A245_226_201; KRTBCOOKIE_80=16514-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&22987-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&23025-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&23386-CAESEMih0bN7ISRdZT8xX8LXzEw; KRTBCOOKIE_377=6810-59dc50c9-d658-44ce-b442-5a1f344d97c0&KRTB&22918-59dc50c9-d658-44ce-b442-5a1f344d97c0&KRTB&23031-59dc50c9-d658-44ce-b442-5a1f344d97c0; uids=eyJ0ZW1wVUlEcyI6eyIzM2Fjcm9zcyI6eyJ1aWQiOiIxMTkxNzkxMDk5Nzc2NjEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTo0My4zODg4Nzk5NVoifSwiYWRmIjp7InVpZCI6IjgwNDQ2MDgzMzM3Nzg4MzkwNzgiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMS4wMzMwNTQ3MjdaIn0sImFka2VybmVsIjp7InVpZCI6IkE5MTYzNTAwNzE0OTkyOTMyOTkwIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuMzczMzg1NjYyWiJ9LCJhZGtlcm5lbEFkbiI6eyJ1aWQiOiJBOTE2MzUwMDcxNDk5MjkzMjk5MCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEzLjQzNDkyNTg5NloifSwiYWRtaXhlciI6eyJ1aWQiOiIzNjZhMTdiMTJmMjI0ZDMwOGYzZTNiOGRhOGMzYzhhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjU5MjkxNDgwMVoifSwiYWRueHMiOnsidWlkIjoiNDE5Mjg5ODUzMDE0NTExOTMiLCJleHBpcmVzIjoiMjAyMy0wMS0xOFQwOTo1MzowOC44MjU0NDI2NzZaIn0sImFqYSI6eyJ1aWQiOiJzMnN1aWQ2RGVmMFl0bjJveGQ1aG9zS1AxVmV3IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTMuMjM5MTc2MDU0WiJ9LCJlcGxhbm5pbmciOnsidWlkIjoiQUoxRjBTOE5qdTdTQ0xWOSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjkyOTk2MDQ3M1oifSwiZ2Ftb3NoaSI6eyJ1aWQiOiJndXNyXzM1NmFmOWIxZDhjNjQyYjQ4MmNiYWQyYjdhMjg4MTYxIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuNTI0MTU3MjI1WiJ9LCJncmlkIjp7InVpZCI6IjRmYzM2MjUwLWQ4NTItNDU5Yy04NzcyLTczNTZkZTE3YWI5NyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE0LjY5NjMxNjIyN1oifSwiZ3JvdXBtIjp7InVpZCI6IjdENzVEMjVGLUZBQzktNDQzRC1CMkQxLUIxN0ZFRTExRTAyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjM5LjIyNjIxMjUzMloifSwiaXgiOnsidWlkIjoiWW9ORlNENlc5QkphOEh6eEdtcXlCUUFBXHUwMDI2Mjk3IiwiZXhwaXJlcyI6IjIwMjMtMDUtMzFUMDc6NTM6MzguNTU1ODI3MzU0WiJ9LCJqaXhpZSI6eyJ1aWQiOiI3MzY3MTI1MC1lODgyLTExZWMtYjUzOC0xM2FjYjdhZjBkZTQiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi4xOTEwOTk3MzJaIn0sImxvZ2ljYWQiOnsidWlkIjoiQVZ4OVROQS11c25pa3M4QURzTHpWa3JvaDg4QUFBR0JUREh0UUEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS40NTUxNDk2MTZaIn0sIm1lZGlhbmV0Ijp7InVpZCI6IjI5Nzg0MjM0OTI4OTU0MTAwMDBWMTAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMy42NzIyMTUxMjhaIn0sIm1naWQiOnsidWlkIjoibTU5Z1hyN0xlX1htIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTcuMDk3MDAxNDcxWiJ9LCJuYW5vaW50ZXJhY3RpdmUiOnsidWlkIjoiNmFlYzhjMTAzNzlkY2I3ODQxMmJjODBiNmRkOWM5NzMxNzNhYjdkNzEyZTQzMWE1YTVlYTcwMzRlNTZhNThhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE2LjcxNDgwNzUwNVoifSwib25ldGFnIjp7InVpZCI6IjdPelZoVzFOeC1LOGFVak1HMG52NXVNYm5YNEFHUXZQbnVHcHFrZ3k0ckEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS4xNDE3NDEyNjJaIn0sIm9wZW54Ijp7InVpZCI6IjVkZWNlNjIyLTBhMjMtMGRhYi0zYTI0LTVhNzcwMTBlNDU4MiIsImV4cGlyZXMiOiIyMDIzLTA1LTMxVDA3OjUyOjQ3LjE0MDQxNzM2M1oifSwicHVibWF0aWMiOnsidWlkIjoiN0Q3NUQyNUYtRkFDOS00NDNELUIyRDEtQjE3RkVFMTFFMDI3IiwiZXhwaXJlcyI6IjIwMjItMTAtMzFUMDk6MTQ6MjUuNzM3MjU2ODk5WiJ9LCJyaWNoYXVkaWVuY2UiOnsidWlkIjoiY2I2YzYzMjAtMzNlMi00Nzc0LWIxNjAtMXp6MTY1NDg0MDc0OSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjUyNTA3NDE4WiJ9LCJzbWFydHlhZHMiOnsidWlkIjoiMTJhZjE1ZTQ0ZjAwZDA3NjMwZTc0YzQ5MTU0Y2JmYmE0Zjg0N2U4ZDRhMTU0YzhjM2Q1MWY1OGNmNzJhNDYyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjgyNTAzMTg4NFoifSwic21pbGV3YW50ZWQiOnsidWlkIjoiZGQ5YzNmZTE4N2VmOWIwOWNhYTViNzExNDA0YzI4MzAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNC4yNTU2MDkzNjNaIn0sInN5bmFjb3JtZWRpYSI6eyJ1aWQiOiJHRFBSIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuOTc5NTgzNDM4WiJ9LCJ0cmlwbGVsaWZ0Ijp7InVpZCI6IjcwMjE5NzUwNTQ4MDg4NjUxOTQ2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA4Ljk4OTY3MzU3NFoifSwidmFsdWVpbXByZXNzaW9uIjp7InVpZCI6IjlkMDgxNTVmLWQ5ZmUtNGI1OC04OThlLWUyYzU2MjgyYWIzZSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjA2NzgzOTE2NFoifSwidmlzeCI6eyJ1aWQiOiIyN2UwYWMzYy1iNDZlLTQxYjMtOTkyYy1mOGQyNzE0OTQ5NWUiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi45ODk1MjM1NzNaIn0sInlpZWxkbGFiIjp7InVpZCI6IjY5NzE0ZDlmLWZiMDAtNGE1Zi04MTljLTRiZTE5MTM2YTMyNSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjExLjMwMzAyNjYxNVoifSwieWllbGRtbyI6eyJ1aWQiOiJnOTZjMmY3MTlmMTU1MWIzMWY2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjExMDUyODYwOVoifSwieWllbGRvbmUiOnsidWlkIjoiMmE0MmZiZDMtMmM3MC00ZWI5LWIxYmQtMDQ2OTY2NTBkOTQ4IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuMzE4MzMzOTM5WiJ9LCJ6ZXJvY2xpY2tmcmF1ZCI6eyJ1aWQiOiJiOTk5NThmZS0yYTg3LTJkYTQtOWNjNC05NjFmZDExM2JlY2UiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNS43MTk1OTQ1NjZaIn19LCJiZGF5IjoiMjAyMi0wNS0xN1QwNjo0ODozOC4wMTc5ODgyMDZaIn0=; KRTBCOOKIE_153=1923-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&19420-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&22979-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&23462-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse; KRTBCOOKIE_57=22776-41928985301451193&KRTB&23339-41928985301451193; KRTBCOOKIE_27=16735-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&16736-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&23019-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&23114-uid:3cab6283-4546-4500-a7b6-40ef605fe745; KRTBCOOKIE_18=22947-1978557989514665832; KRTBCOOKIE_466=16530-4fc36250-d852-459c-8772-7356de17ab97; KRTBCOOKIE_391=22924-8044608333778839078&KRTB&23263-8044608333778839078&KRTB&23481-8044608333778839078; KRTBCOOKIE_1310=23431-b81c3g7dr67i&KRTB&23446-b81c3g7dr67i&KRTB&23465-b81c3g7dr67i; KRTBCOOKIE_1290=23368-vkf3yv9lbbl; KRTBCOOKIE_22=14911-4554572065121110164&KRTB&23150-4554572065121110164; KRTBCOOKIE_860=16335-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23334-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23417-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23426-YGAqDU1zUTdjyAFxCoe3kctlNPo; KRTBCOOKIE_904=16787-KwJwE7NkCZClNJRysN2iYg; KRTBCOOKIE_1159=23138-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23328-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23427-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23445-5545f53f3d6e4ec199d8ed627ff026f3; KRTBCOOKIE_32=11175-AQEI_1QecY2ESAIjEW6KAQEBAQE&KRTB&22713-AQEI_1QecY2ESAIjEW6KAQEBAQE&KRTB&22715-AQEI_1QecY2ESAIjEW6KAQEBAQE; SyncRTB3=1685577600%3A35%7C1685491200%3A107_21_71_56_204_247_165_231_233_179_22_209_54_254_238_96_99_220_7_214_13_3_8_234_176_46_5%7C1684886400%3A2_223_15%7C1689465600%3A69%7C1685145600%3A63; KRTBCOOKIE_107=1471-uid:EK38R0PM1NQR0H5&KRTB&23421-uid:EK38R0PM1NQR0H5; KRTBCOOKIE_594=17105-RX-447a6332-530e-456a-97f4-3f0fd1ed48c9-004&KRTB&17107-RX-447a6332-530e-456a-97f4-3f0fd1ed48c9-004; SPugT=1684310122; chkChromeAb67Sec=133; KRTBCOOKIE_699=22727-AAFy2k7FBosAAEasbJoXnw; PugT=1684310473; origin=go-test`)
						return r
					}(),
					Body: []byte(`{"ext":{"wrapper":{"profileid":5890,"versionid":1}},"app":{"publisher":{"id":"5890"}}}`),
				},
				setup: func(mme *mock_metrics.MockMetricsEngine) {},
			},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{
				ModuleContext: hookstage.ModuleContext{
					"rctx": models.RequestCtx{
						PubIDStr:                  "5890",
						PubID:                     5890,
						ProfileID:                 5890,
						DisplayID:                 1,
						DisplayVersionID:          1,
						SSAuction:                 -1,
						Debug:                     true,
						DeviceCtx:                 models.DeviceCtx{IP: "127.0.0.1", UA: "go-test"},
						IsCTVRequest:              false,
						TrackerEndpoint:           "t.pubmatic.com",
						VideoErrorTrackerEndpoint: "t.pubmatic.com/error",
						UidCookie: &http.Cookie{
							Name:  "uids",
							Value: `eyJ0ZW1wVUlEcyI6eyIzM2Fjcm9zcyI6eyJ1aWQiOiIxMTkxNzkxMDk5Nzc2NjEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTo0My4zODg4Nzk5NVoifSwiYWRmIjp7InVpZCI6IjgwNDQ2MDgzMzM3Nzg4MzkwNzgiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMS4wMzMwNTQ3MjdaIn0sImFka2VybmVsIjp7InVpZCI6IkE5MTYzNTAwNzE0OTkyOTMyOTkwIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuMzczMzg1NjYyWiJ9LCJhZGtlcm5lbEFkbiI6eyJ1aWQiOiJBOTE2MzUwMDcxNDk5MjkzMjk5MCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEzLjQzNDkyNTg5NloifSwiYWRtaXhlciI6eyJ1aWQiOiIzNjZhMTdiMTJmMjI0ZDMwOGYzZTNiOGRhOGMzYzhhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjU5MjkxNDgwMVoifSwiYWRueHMiOnsidWlkIjoiNDE5Mjg5ODUzMDE0NTExOTMiLCJleHBpcmVzIjoiMjAyMy0wMS0xOFQwOTo1MzowOC44MjU0NDI2NzZaIn0sImFqYSI6eyJ1aWQiOiJzMnN1aWQ2RGVmMFl0bjJveGQ1aG9zS1AxVmV3IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTMuMjM5MTc2MDU0WiJ9LCJlcGxhbm5pbmciOnsidWlkIjoiQUoxRjBTOE5qdTdTQ0xWOSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjkyOTk2MDQ3M1oifSwiZ2Ftb3NoaSI6eyJ1aWQiOiJndXNyXzM1NmFmOWIxZDhjNjQyYjQ4MmNiYWQyYjdhMjg4MTYxIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuNTI0MTU3MjI1WiJ9LCJncmlkIjp7InVpZCI6IjRmYzM2MjUwLWQ4NTItNDU5Yy04NzcyLTczNTZkZTE3YWI5NyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE0LjY5NjMxNjIyN1oifSwiZ3JvdXBtIjp7InVpZCI6IjdENzVEMjVGLUZBQzktNDQzRC1CMkQxLUIxN0ZFRTExRTAyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjM5LjIyNjIxMjUzMloifSwiaXgiOnsidWlkIjoiWW9ORlNENlc5QkphOEh6eEdtcXlCUUFBXHUwMDI2Mjk3IiwiZXhwaXJlcyI6IjIwMjMtMDUtMzFUMDc6NTM6MzguNTU1ODI3MzU0WiJ9LCJqaXhpZSI6eyJ1aWQiOiI3MzY3MTI1MC1lODgyLTExZWMtYjUzOC0xM2FjYjdhZjBkZTQiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi4xOTEwOTk3MzJaIn0sImxvZ2ljYWQiOnsidWlkIjoiQVZ4OVROQS11c25pa3M4QURzTHpWa3JvaDg4QUFBR0JUREh0UUEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS40NTUxNDk2MTZaIn0sIm1lZGlhbmV0Ijp7InVpZCI6IjI5Nzg0MjM0OTI4OTU0MTAwMDBWMTAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMy42NzIyMTUxMjhaIn0sIm1naWQiOnsidWlkIjoibTU5Z1hyN0xlX1htIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTcuMDk3MDAxNDcxWiJ9LCJuYW5vaW50ZXJhY3RpdmUiOnsidWlkIjoiNmFlYzhjMTAzNzlkY2I3ODQxMmJjODBiNmRkOWM5NzMxNzNhYjdkNzEyZTQzMWE1YTVlYTcwMzRlNTZhNThhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE2LjcxNDgwNzUwNVoifSwib25ldGFnIjp7InVpZCI6IjdPelZoVzFOeC1LOGFVak1HMG52NXVNYm5YNEFHUXZQbnVHcHFrZ3k0ckEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS4xNDE3NDEyNjJaIn0sIm9wZW54Ijp7InVpZCI6IjVkZWNlNjIyLTBhMjMtMGRhYi0zYTI0LTVhNzcwMTBlNDU4MiIsImV4cGlyZXMiOiIyMDIzLTA1LTMxVDA3OjUyOjQ3LjE0MDQxNzM2M1oifSwicHVibWF0aWMiOnsidWlkIjoiN0Q3NUQyNUYtRkFDOS00NDNELUIyRDEtQjE3RkVFMTFFMDI3IiwiZXhwaXJlcyI6IjIwMjItMTAtMzFUMDk6MTQ6MjUuNzM3MjU2ODk5WiJ9LCJyaWNoYXVkaWVuY2UiOnsidWlkIjoiY2I2YzYzMjAtMzNlMi00Nzc0LWIxNjAtMXp6MTY1NDg0MDc0OSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjUyNTA3NDE4WiJ9LCJzbWFydHlhZHMiOnsidWlkIjoiMTJhZjE1ZTQ0ZjAwZDA3NjMwZTc0YzQ5MTU0Y2JmYmE0Zjg0N2U4ZDRhMTU0YzhjM2Q1MWY1OGNmNzJhNDYyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjgyNTAzMTg4NFoifSwic21pbGV3YW50ZWQiOnsidWlkIjoiZGQ5YzNmZTE4N2VmOWIwOWNhYTViNzExNDA0YzI4MzAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNC4yNTU2MDkzNjNaIn0sInN5bmFjb3JtZWRpYSI6eyJ1aWQiOiJHRFBSIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuOTc5NTgzNDM4WiJ9LCJ0cmlwbGVsaWZ0Ijp7InVpZCI6IjcwMjE5NzUwNTQ4MDg4NjUxOTQ2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA4Ljk4OTY3MzU3NFoifSwidmFsdWVpbXByZXNzaW9uIjp7InVpZCI6IjlkMDgxNTVmLWQ5ZmUtNGI1OC04OThlLWUyYzU2MjgyYWIzZSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjA2NzgzOTE2NFoifSwidmlzeCI6eyJ1aWQiOiIyN2UwYWMzYy1iNDZlLTQxYjMtOTkyYy1mOGQyNzE0OTQ5NWUiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi45ODk1MjM1NzNaIn0sInlpZWxkbGFiIjp7InVpZCI6IjY5NzE0ZDlmLWZiMDAtNGE1Zi04MTljLTRiZTE5MTM2YTMyNSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjExLjMwMzAyNjYxNVoifSwieWllbGRtbyI6eyJ1aWQiOiJnOTZjMmY3MTlmMTU1MWIzMWY2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjExMDUyODYwOVoifSwieWllbGRvbmUiOnsidWlkIjoiMmE0MmZiZDMtMmM3MC00ZWI5LWIxYmQtMDQ2OTY2NTBkOTQ4IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuMzE4MzMzOTM5WiJ9LCJ6ZXJvY2xpY2tmcmF1ZCI6eyJ1aWQiOiJiOTk5NThmZS0yYTg3LTJkYTQtOWNjNC05NjFmZDExM2JlY2UiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNS43MTk1OTQ1NjZaIn19LCJiZGF5IjoiMjAyMi0wNS0xN1QwNjo0ODozOC4wMTc5ODgyMDZaIn0=`,
						},
						KADUSERCookie: &http.Cookie{
							Name:  "KADUSERCOOKIE",
							Value: `7D75D25F-FAC9-443D-B2D1-B17FEE11E027`,
						},
						OriginCookie:                    "go-test",
						Aliases:                         make(map[string]string),
						ImpBidCtx:                       make(map[string]models.ImpCtx),
						PrebidBidderCode:                make(map[string]string),
						BidderResponseTimeMillis:        make(map[string]int),
						ProfileIDStr:                    "5890",
						Endpoint:                        models.EndpointV25,
						MetricsEngine:                   mockEngine,
						SeatNonBids:                     make(map[string][]openrtb_ext.NonBid),
						Method:                          "POST",
						WakandaDebug:                    &wakanda.Debug{},
						ImpCountingMethodEnabledBidders: make(map[string]struct{}),
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "valid /openrtb/2.5 request with wiid set and no cookies(reverseProxy through SSHB(8001->8000)",
			fields: fields{
				cfg: config.Config{
					Tracker: config.Tracker{
						Endpoint:                  "t.pubmatic.com",
						VideoErrorTrackerEndpoint: "t.pubmatic.com/error",
					},
				},
				cache: nil,
			},
			args: args{
				in0:   context.Background(),
				miCtx: hookstage.ModuleInvocationContext{},
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						r, err := http.NewRequest("POST", "http://localhost/openrtb/2.5?debug=1&sshb=2", nil)
						if err != nil {
							panic(err)
						}
						r.Header.Add("User-Agent", "go-test")
						r.Header.Add("SOURCE_IP", "127.0.0.1")
						return r
					}(),
					Body: []byte(`{"ext":{"wrapper":{"profileid":5890,"versionid":1,"wiid":"4df09505-d0b2-4d70-94d9-dc41e8e777f7"}},"app":{"publisher":{"id":"5890"}}}`),
				},
				setup: func(mme *mock_metrics.MockMetricsEngine) {},
			},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{
				ModuleContext: hookstage.ModuleContext{
					"rctx": models.RequestCtx{
						PubIDStr:                        "5890",
						PubID:                           5890,
						ProfileID:                       5890,
						DisplayID:                       1,
						DisplayVersionID:                1,
						SSAuction:                       -1,
						Debug:                           true,
						DeviceCtx:                       models.DeviceCtx{IP: "127.0.0.1", UA: "go-test"},
						IsCTVRequest:                    false,
						TrackerEndpoint:                 "t.pubmatic.com",
						VideoErrorTrackerEndpoint:       "t.pubmatic.com/error",
						LoggerImpressionID:              "4df09505-d0b2-4d70-94d9-dc41e8e777f7",
						Aliases:                         make(map[string]string),
						ImpBidCtx:                       make(map[string]models.ImpCtx),
						PrebidBidderCode:                make(map[string]string),
						BidderResponseTimeMillis:        make(map[string]int),
						ProfileIDStr:                    "5890",
						Endpoint:                        models.EndpointV25,
						MetricsEngine:                   mockEngine,
						SeatNonBids:                     make(map[string][]openrtb_ext.NonBid),
						Method:                          "POST",
						WakandaDebug:                    &wakanda.Debug{},
						ImpCountingMethodEnabledBidders: make(map[string]struct{}),
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "/openrtb/2.5 request without profileid(reverseProxy through SSHB(8001->8000)",
			fields: fields{
				cfg:   config.Config{},
				cache: nil,
			},
			args: args{
				in0:   context.Background(),
				miCtx: hookstage.ModuleInvocationContext{},
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						r, err := http.NewRequest("POST", "http://localhost/openrtb/2.5?&sshb=2", nil)
						if err != nil {
							panic(err)
						}
						return r
					}(),
					Body: []byte(`{"ext":{"wrapper":{"profileids":5890,"versionid":1}}}`),
				},
				setup: func(mme *mock_metrics.MockMetricsEngine) {
					mme.EXPECT().RecordBadRequests(gomock.Any(), "0", 700)
				},
			},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{
				Reject:  true,
				NbrCode: int(nbr.InvalidProfileID),
				Errors:  []string{"ErrMissingProfileID"},
			},
			wantErr: nil,
		},
		{
			name: "Valid /openrtb2/auction?source=pbjs request(webs2s)",
			fields: fields{
				cfg:   config.Config{},
				cache: nil,
			},
			args: args{
				in0:   context.Background(),
				miCtx: hookstage.ModuleInvocationContext{},
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						r, err := http.NewRequest("POST", "http://localhost/openrtb2/auction?source=pbjs&debug=1", nil)
						if err != nil {
							panic(err)
						}
						r.Header.Add("User-Agent", "go-test")
						r.Header.Add("SOURCE_IP", "127.0.0.1")
						r.Header.Add("Cookie", `KADUSERCOOKIE=7D75D25F-FAC9-443D-B2D1-B17FEE11E027; DPSync3=1684886400%3A248%7C1685491200%3A245_226_201; KRTBCOOKIE_80=16514-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&22987-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&23025-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&23386-CAESEMih0bN7ISRdZT8xX8LXzEw; KRTBCOOKIE_377=6810-59dc50c9-d658-44ce-b442-5a1f344d97c0&KRTB&22918-59dc50c9-d658-44ce-b442-5a1f344d97c0&KRTB&23031-59dc50c9-d658-44ce-b442-5a1f344d97c0; uids=eyJ0ZW1wVUlEcyI6eyIzM2Fjcm9zcyI6eyJ1aWQiOiIxMTkxNzkxMDk5Nzc2NjEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTo0My4zODg4Nzk5NVoifSwiYWRmIjp7InVpZCI6IjgwNDQ2MDgzMzM3Nzg4MzkwNzgiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMS4wMzMwNTQ3MjdaIn0sImFka2VybmVsIjp7InVpZCI6IkE5MTYzNTAwNzE0OTkyOTMyOTkwIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuMzczMzg1NjYyWiJ9LCJhZGtlcm5lbEFkbiI6eyJ1aWQiOiJBOTE2MzUwMDcxNDk5MjkzMjk5MCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEzLjQzNDkyNTg5NloifSwiYWRtaXhlciI6eyJ1aWQiOiIzNjZhMTdiMTJmMjI0ZDMwOGYzZTNiOGRhOGMzYzhhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjU5MjkxNDgwMVoifSwiYWRueHMiOnsidWlkIjoiNDE5Mjg5ODUzMDE0NTExOTMiLCJleHBpcmVzIjoiMjAyMy0wMS0xOFQwOTo1MzowOC44MjU0NDI2NzZaIn0sImFqYSI6eyJ1aWQiOiJzMnN1aWQ2RGVmMFl0bjJveGQ1aG9zS1AxVmV3IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTMuMjM5MTc2MDU0WiJ9LCJlcGxhbm5pbmciOnsidWlkIjoiQUoxRjBTOE5qdTdTQ0xWOSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjkyOTk2MDQ3M1oifSwiZ2Ftb3NoaSI6eyJ1aWQiOiJndXNyXzM1NmFmOWIxZDhjNjQyYjQ4MmNiYWQyYjdhMjg4MTYxIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuNTI0MTU3MjI1WiJ9LCJncmlkIjp7InVpZCI6IjRmYzM2MjUwLWQ4NTItNDU5Yy04NzcyLTczNTZkZTE3YWI5NyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE0LjY5NjMxNjIyN1oifSwiZ3JvdXBtIjp7InVpZCI6IjdENzVEMjVGLUZBQzktNDQzRC1CMkQxLUIxN0ZFRTExRTAyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjM5LjIyNjIxMjUzMloifSwiaXgiOnsidWlkIjoiWW9ORlNENlc5QkphOEh6eEdtcXlCUUFBXHUwMDI2Mjk3IiwiZXhwaXJlcyI6IjIwMjMtMDUtMzFUMDc6NTM6MzguNTU1ODI3MzU0WiJ9LCJqaXhpZSI6eyJ1aWQiOiI3MzY3MTI1MC1lODgyLTExZWMtYjUzOC0xM2FjYjdhZjBkZTQiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi4xOTEwOTk3MzJaIn0sImxvZ2ljYWQiOnsidWlkIjoiQVZ4OVROQS11c25pa3M4QURzTHpWa3JvaDg4QUFBR0JUREh0UUEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS40NTUxNDk2MTZaIn0sIm1lZGlhbmV0Ijp7InVpZCI6IjI5Nzg0MjM0OTI4OTU0MTAwMDBWMTAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMy42NzIyMTUxMjhaIn0sIm1naWQiOnsidWlkIjoibTU5Z1hyN0xlX1htIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTcuMDk3MDAxNDcxWiJ9LCJuYW5vaW50ZXJhY3RpdmUiOnsidWlkIjoiNmFlYzhjMTAzNzlkY2I3ODQxMmJjODBiNmRkOWM5NzMxNzNhYjdkNzEyZTQzMWE1YTVlYTcwMzRlNTZhNThhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE2LjcxNDgwNzUwNVoifSwib25ldGFnIjp7InVpZCI6IjdPelZoVzFOeC1LOGFVak1HMG52NXVNYm5YNEFHUXZQbnVHcHFrZ3k0ckEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS4xNDE3NDEyNjJaIn0sIm9wZW54Ijp7InVpZCI6IjVkZWNlNjIyLTBhMjMtMGRhYi0zYTI0LTVhNzcwMTBlNDU4MiIsImV4cGlyZXMiOiIyMDIzLTA1LTMxVDA3OjUyOjQ3LjE0MDQxNzM2M1oifSwicHVibWF0aWMiOnsidWlkIjoiN0Q3NUQyNUYtRkFDOS00NDNELUIyRDEtQjE3RkVFMTFFMDI3IiwiZXhwaXJlcyI6IjIwMjItMTAtMzFUMDk6MTQ6MjUuNzM3MjU2ODk5WiJ9LCJyaWNoYXVkaWVuY2UiOnsidWlkIjoiY2I2YzYzMjAtMzNlMi00Nzc0LWIxNjAtMXp6MTY1NDg0MDc0OSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjUyNTA3NDE4WiJ9LCJzbWFydHlhZHMiOnsidWlkIjoiMTJhZjE1ZTQ0ZjAwZDA3NjMwZTc0YzQ5MTU0Y2JmYmE0Zjg0N2U4ZDRhMTU0YzhjM2Q1MWY1OGNmNzJhNDYyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjgyNTAzMTg4NFoifSwic21pbGV3YW50ZWQiOnsidWlkIjoiZGQ5YzNmZTE4N2VmOWIwOWNhYTViNzExNDA0YzI4MzAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNC4yNTU2MDkzNjNaIn0sInN5bmFjb3JtZWRpYSI6eyJ1aWQiOiJHRFBSIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuOTc5NTgzNDM4WiJ9LCJ0cmlwbGVsaWZ0Ijp7InVpZCI6IjcwMjE5NzUwNTQ4MDg4NjUxOTQ2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA4Ljk4OTY3MzU3NFoifSwidmFsdWVpbXByZXNzaW9uIjp7InVpZCI6IjlkMDgxNTVmLWQ5ZmUtNGI1OC04OThlLWUyYzU2MjgyYWIzZSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjA2NzgzOTE2NFoifSwidmlzeCI6eyJ1aWQiOiIyN2UwYWMzYy1iNDZlLTQxYjMtOTkyYy1mOGQyNzE0OTQ5NWUiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi45ODk1MjM1NzNaIn0sInlpZWxkbGFiIjp7InVpZCI6IjY5NzE0ZDlmLWZiMDAtNGE1Zi04MTljLTRiZTE5MTM2YTMyNSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjExLjMwMzAyNjYxNVoifSwieWllbGRtbyI6eyJ1aWQiOiJnOTZjMmY3MTlmMTU1MWIzMWY2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjExMDUyODYwOVoifSwieWllbGRvbmUiOnsidWlkIjoiMmE0MmZiZDMtMmM3MC00ZWI5LWIxYmQtMDQ2OTY2NTBkOTQ4IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuMzE4MzMzOTM5WiJ9LCJ6ZXJvY2xpY2tmcmF1ZCI6eyJ1aWQiOiJiOTk5NThmZS0yYTg3LTJkYTQtOWNjNC05NjFmZDExM2JlY2UiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNS43MTk1OTQ1NjZaIn19LCJiZGF5IjoiMjAyMi0wNS0xN1QwNjo0ODozOC4wMTc5ODgyMDZaIn0=; KRTBCOOKIE_153=1923-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&19420-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&22979-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&23462-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse; KRTBCOOKIE_57=22776-41928985301451193&KRTB&23339-41928985301451193; KRTBCOOKIE_27=16735-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&16736-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&23019-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&23114-uid:3cab6283-4546-4500-a7b6-40ef605fe745; KRTBCOOKIE_18=22947-1978557989514665832; KRTBCOOKIE_466=16530-4fc36250-d852-459c-8772-7356de17ab97; KRTBCOOKIE_391=22924-8044608333778839078&KRTB&23263-8044608333778839078&KRTB&23481-8044608333778839078; KRTBCOOKIE_1310=23431-b81c3g7dr67i&KRTB&23446-b81c3g7dr67i&KRTB&23465-b81c3g7dr67i; KRTBCOOKIE_1290=23368-vkf3yv9lbbl; KRTBCOOKIE_22=14911-4554572065121110164&KRTB&23150-4554572065121110164; KRTBCOOKIE_860=16335-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23334-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23417-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23426-YGAqDU1zUTdjyAFxCoe3kctlNPo; KRTBCOOKIE_904=16787-KwJwE7NkCZClNJRysN2iYg; KRTBCOOKIE_1159=23138-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23328-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23427-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23445-5545f53f3d6e4ec199d8ed627ff026f3; KRTBCOOKIE_32=11175-AQEI_1QecY2ESAIjEW6KAQEBAQE&KRTB&22713-AQEI_1QecY2ESAIjEW6KAQEBAQE&KRTB&22715-AQEI_1QecY2ESAIjEW6KAQEBAQE; SyncRTB3=1685577600%3A35%7C1685491200%3A107_21_71_56_204_247_165_231_233_179_22_209_54_254_238_96_99_220_7_214_13_3_8_234_176_46_5%7C1684886400%3A2_223_15%7C1689465600%3A69%7C1685145600%3A63; KRTBCOOKIE_107=1471-uid:EK38R0PM1NQR0H5&KRTB&23421-uid:EK38R0PM1NQR0H5; KRTBCOOKIE_594=17105-RX-447a6332-530e-456a-97f4-3f0fd1ed48c9-004&KRTB&17107-RX-447a6332-530e-456a-97f4-3f0fd1ed48c9-004; SPugT=1684310122; chkChromeAb67Sec=133; KRTBCOOKIE_699=22727-AAFy2k7FBosAAEasbJoXnw; PugT=1684310473; origin=go-test`)
						return r
					}(),
					Body: []byte(`{"imp":[{"tagid":"/43743431/DMDemo","ext":{"prebid":{"bidder":{"pubmatic":{"publisherId":"5890"}},"adunitcode":"div-gpt-ad-1460505748561-0"}},"id":"div-gpt-ad-1460505748561-0","banner":{"topframe":1,"format":[{"w":300,"h":250}]}}],"site":{"domain":"localhost:9999","publisher":{"domain":"localhost:9999","id":"5890"},"page":"http://localhost:9999/integrationExamples/gpt/owServer_example.html"},"device":{"w":1792,"h":446,"dnt":0,"ua":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36","language":"en","sua":{"source":1,"platform":{"brand":"macOS"},"browsers":[{"brand":"Google Chrome","version":["117"]},{"brand":"Not;A=Brand","version":["8"]},{"brand":"Chromium","version":["117"]}],"mobile":0}},"ext":{"prebid":{"auctiontimestamp":1697191822565,"targeting":{"includewinners":true,"includebidderkeys":true},"bidderparams":{"pubmatic":{"publisherId":"5890","wrapper":{"profileid":43563,"versionid":1}}},"channel":{"name":"pbjs","version":"v8.7.0-pre"},"createtids":false}},"id":"5bdd7da5-1166-40fe-a9cb-3bf3c3164cd3","test":0,"tmax":3000}`),
				},
				setup: func(mme *mock_metrics.MockMetricsEngine) {},
			},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{
				ModuleContext: hookstage.ModuleContext{
					"rctx": models.RequestCtx{
						ProfileID:        43563,
						PubID:            5890,
						PubIDStr:         "5890",
						DisplayID:        1,
						DisplayVersionID: 1,
						SSAuction:        -1,
						Debug:            true,
						DeviceCtx:        models.DeviceCtx{IP: "127.0.0.1", UA: "go-test"},
						IsCTVRequest:     false,
						UidCookie: &http.Cookie{
							Name:  "uids",
							Value: `eyJ0ZW1wVUlEcyI6eyIzM2Fjcm9zcyI6eyJ1aWQiOiIxMTkxNzkxMDk5Nzc2NjEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTo0My4zODg4Nzk5NVoifSwiYWRmIjp7InVpZCI6IjgwNDQ2MDgzMzM3Nzg4MzkwNzgiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMS4wMzMwNTQ3MjdaIn0sImFka2VybmVsIjp7InVpZCI6IkE5MTYzNTAwNzE0OTkyOTMyOTkwIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuMzczMzg1NjYyWiJ9LCJhZGtlcm5lbEFkbiI6eyJ1aWQiOiJBOTE2MzUwMDcxNDk5MjkzMjk5MCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEzLjQzNDkyNTg5NloifSwiYWRtaXhlciI6eyJ1aWQiOiIzNjZhMTdiMTJmMjI0ZDMwOGYzZTNiOGRhOGMzYzhhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjU5MjkxNDgwMVoifSwiYWRueHMiOnsidWlkIjoiNDE5Mjg5ODUzMDE0NTExOTMiLCJleHBpcmVzIjoiMjAyMy0wMS0xOFQwOTo1MzowOC44MjU0NDI2NzZaIn0sImFqYSI6eyJ1aWQiOiJzMnN1aWQ2RGVmMFl0bjJveGQ1aG9zS1AxVmV3IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTMuMjM5MTc2MDU0WiJ9LCJlcGxhbm5pbmciOnsidWlkIjoiQUoxRjBTOE5qdTdTQ0xWOSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjkyOTk2MDQ3M1oifSwiZ2Ftb3NoaSI6eyJ1aWQiOiJndXNyXzM1NmFmOWIxZDhjNjQyYjQ4MmNiYWQyYjdhMjg4MTYxIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuNTI0MTU3MjI1WiJ9LCJncmlkIjp7InVpZCI6IjRmYzM2MjUwLWQ4NTItNDU5Yy04NzcyLTczNTZkZTE3YWI5NyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE0LjY5NjMxNjIyN1oifSwiZ3JvdXBtIjp7InVpZCI6IjdENzVEMjVGLUZBQzktNDQzRC1CMkQxLUIxN0ZFRTExRTAyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjM5LjIyNjIxMjUzMloifSwiaXgiOnsidWlkIjoiWW9ORlNENlc5QkphOEh6eEdtcXlCUUFBXHUwMDI2Mjk3IiwiZXhwaXJlcyI6IjIwMjMtMDUtMzFUMDc6NTM6MzguNTU1ODI3MzU0WiJ9LCJqaXhpZSI6eyJ1aWQiOiI3MzY3MTI1MC1lODgyLTExZWMtYjUzOC0xM2FjYjdhZjBkZTQiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi4xOTEwOTk3MzJaIn0sImxvZ2ljYWQiOnsidWlkIjoiQVZ4OVROQS11c25pa3M4QURzTHpWa3JvaDg4QUFBR0JUREh0UUEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS40NTUxNDk2MTZaIn0sIm1lZGlhbmV0Ijp7InVpZCI6IjI5Nzg0MjM0OTI4OTU0MTAwMDBWMTAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMy42NzIyMTUxMjhaIn0sIm1naWQiOnsidWlkIjoibTU5Z1hyN0xlX1htIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTcuMDk3MDAxNDcxWiJ9LCJuYW5vaW50ZXJhY3RpdmUiOnsidWlkIjoiNmFlYzhjMTAzNzlkY2I3ODQxMmJjODBiNmRkOWM5NzMxNzNhYjdkNzEyZTQzMWE1YTVlYTcwMzRlNTZhNThhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE2LjcxNDgwNzUwNVoifSwib25ldGFnIjp7InVpZCI6IjdPelZoVzFOeC1LOGFVak1HMG52NXVNYm5YNEFHUXZQbnVHcHFrZ3k0ckEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS4xNDE3NDEyNjJaIn0sIm9wZW54Ijp7InVpZCI6IjVkZWNlNjIyLTBhMjMtMGRhYi0zYTI0LTVhNzcwMTBlNDU4MiIsImV4cGlyZXMiOiIyMDIzLTA1LTMxVDA3OjUyOjQ3LjE0MDQxNzM2M1oifSwicHVibWF0aWMiOnsidWlkIjoiN0Q3NUQyNUYtRkFDOS00NDNELUIyRDEtQjE3RkVFMTFFMDI3IiwiZXhwaXJlcyI6IjIwMjItMTAtMzFUMDk6MTQ6MjUuNzM3MjU2ODk5WiJ9LCJyaWNoYXVkaWVuY2UiOnsidWlkIjoiY2I2YzYzMjAtMzNlMi00Nzc0LWIxNjAtMXp6MTY1NDg0MDc0OSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjUyNTA3NDE4WiJ9LCJzbWFydHlhZHMiOnsidWlkIjoiMTJhZjE1ZTQ0ZjAwZDA3NjMwZTc0YzQ5MTU0Y2JmYmE0Zjg0N2U4ZDRhMTU0YzhjM2Q1MWY1OGNmNzJhNDYyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjgyNTAzMTg4NFoifSwic21pbGV3YW50ZWQiOnsidWlkIjoiZGQ5YzNmZTE4N2VmOWIwOWNhYTViNzExNDA0YzI4MzAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNC4yNTU2MDkzNjNaIn0sInN5bmFjb3JtZWRpYSI6eyJ1aWQiOiJHRFBSIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuOTc5NTgzNDM4WiJ9LCJ0cmlwbGVsaWZ0Ijp7InVpZCI6IjcwMjE5NzUwNTQ4MDg4NjUxOTQ2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA4Ljk4OTY3MzU3NFoifSwidmFsdWVpbXByZXNzaW9uIjp7InVpZCI6IjlkMDgxNTVmLWQ5ZmUtNGI1OC04OThlLWUyYzU2MjgyYWIzZSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjA2NzgzOTE2NFoifSwidmlzeCI6eyJ1aWQiOiIyN2UwYWMzYy1iNDZlLTQxYjMtOTkyYy1mOGQyNzE0OTQ5NWUiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi45ODk1MjM1NzNaIn0sInlpZWxkbGFiIjp7InVpZCI6IjY5NzE0ZDlmLWZiMDAtNGE1Zi04MTljLTRiZTE5MTM2YTMyNSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjExLjMwMzAyNjYxNVoifSwieWllbGRtbyI6eyJ1aWQiOiJnOTZjMmY3MTlmMTU1MWIzMWY2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjExMDUyODYwOVoifSwieWllbGRvbmUiOnsidWlkIjoiMmE0MmZiZDMtMmM3MC00ZWI5LWIxYmQtMDQ2OTY2NTBkOTQ4IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuMzE4MzMzOTM5WiJ9LCJ6ZXJvY2xpY2tmcmF1ZCI6eyJ1aWQiOiJiOTk5NThmZS0yYTg3LTJkYTQtOWNjNC05NjFmZDExM2JlY2UiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNS43MTk1OTQ1NjZaIn19LCJiZGF5IjoiMjAyMi0wNS0xN1QwNjo0ODozOC4wMTc5ODgyMDZaIn0=`,
						},
						KADUSERCookie: &http.Cookie{
							Name:  "KADUSERCOOKIE",
							Value: `7D75D25F-FAC9-443D-B2D1-B17FEE11E027`,
						},
						OriginCookie:                    "go-test",
						Aliases:                         make(map[string]string),
						ImpBidCtx:                       make(map[string]models.ImpCtx),
						PrebidBidderCode:                make(map[string]string),
						BidderResponseTimeMillis:        make(map[string]int),
						ProfileIDStr:                    "43563",
						Endpoint:                        models.EndpointWebS2S,
						MetricsEngine:                   mockEngine,
						SeatNonBids:                     make(map[string][]openrtb_ext.NonBid),
						Method:                          "POST",
						WakandaDebug:                    &wakanda.Debug{},
						ImpCountingMethodEnabledBidders: make(map[string]struct{}),
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "Valid /openrtb2/auction?source=pbjs request(webs2s) debug set from request body instead of queryparam",
			fields: fields{
				cfg:   config.Config{},
				cache: nil,
			},
			args: args{
				in0:   context.Background(),
				miCtx: hookstage.ModuleInvocationContext{},
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						r, err := http.NewRequest("POST", "http://localhost/openrtb2/auction?source=pbjs", nil)
						if err != nil {
							panic(err)
						}
						r.Header.Add("User-Agent", "go-test")
						r.Header.Add("SOURCE_IP", "127.0.0.1")
						r.Header.Add("Cookie", `KADUSERCOOKIE=7D75D25F-FAC9-443D-B2D1-B17FEE11E027; DPSync3=1684886400%3A248%7C1685491200%3A245_226_201; KRTBCOOKIE_80=16514-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&22987-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&23025-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&23386-CAESEMih0bN7ISRdZT8xX8LXzEw; KRTBCOOKIE_377=6810-59dc50c9-d658-44ce-b442-5a1f344d97c0&KRTB&22918-59dc50c9-d658-44ce-b442-5a1f344d97c0&KRTB&23031-59dc50c9-d658-44ce-b442-5a1f344d97c0; uids=eyJ0ZW1wVUlEcyI6eyIzM2Fjcm9zcyI6eyJ1aWQiOiIxMTkxNzkxMDk5Nzc2NjEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTo0My4zODg4Nzk5NVoifSwiYWRmIjp7InVpZCI6IjgwNDQ2MDgzMzM3Nzg4MzkwNzgiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMS4wMzMwNTQ3MjdaIn0sImFka2VybmVsIjp7InVpZCI6IkE5MTYzNTAwNzE0OTkyOTMyOTkwIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuMzczMzg1NjYyWiJ9LCJhZGtlcm5lbEFkbiI6eyJ1aWQiOiJBOTE2MzUwMDcxNDk5MjkzMjk5MCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEzLjQzNDkyNTg5NloifSwiYWRtaXhlciI6eyJ1aWQiOiIzNjZhMTdiMTJmMjI0ZDMwOGYzZTNiOGRhOGMzYzhhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjU5MjkxNDgwMVoifSwiYWRueHMiOnsidWlkIjoiNDE5Mjg5ODUzMDE0NTExOTMiLCJleHBpcmVzIjoiMjAyMy0wMS0xOFQwOTo1MzowOC44MjU0NDI2NzZaIn0sImFqYSI6eyJ1aWQiOiJzMnN1aWQ2RGVmMFl0bjJveGQ1aG9zS1AxVmV3IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTMuMjM5MTc2MDU0WiJ9LCJlcGxhbm5pbmciOnsidWlkIjoiQUoxRjBTOE5qdTdTQ0xWOSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjkyOTk2MDQ3M1oifSwiZ2Ftb3NoaSI6eyJ1aWQiOiJndXNyXzM1NmFmOWIxZDhjNjQyYjQ4MmNiYWQyYjdhMjg4MTYxIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuNTI0MTU3MjI1WiJ9LCJncmlkIjp7InVpZCI6IjRmYzM2MjUwLWQ4NTItNDU5Yy04NzcyLTczNTZkZTE3YWI5NyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE0LjY5NjMxNjIyN1oifSwiZ3JvdXBtIjp7InVpZCI6IjdENzVEMjVGLUZBQzktNDQzRC1CMkQxLUIxN0ZFRTExRTAyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjM5LjIyNjIxMjUzMloifSwiaXgiOnsidWlkIjoiWW9ORlNENlc5QkphOEh6eEdtcXlCUUFBXHUwMDI2Mjk3IiwiZXhwaXJlcyI6IjIwMjMtMDUtMzFUMDc6NTM6MzguNTU1ODI3MzU0WiJ9LCJqaXhpZSI6eyJ1aWQiOiI3MzY3MTI1MC1lODgyLTExZWMtYjUzOC0xM2FjYjdhZjBkZTQiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi4xOTEwOTk3MzJaIn0sImxvZ2ljYWQiOnsidWlkIjoiQVZ4OVROQS11c25pa3M4QURzTHpWa3JvaDg4QUFBR0JUREh0UUEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS40NTUxNDk2MTZaIn0sIm1lZGlhbmV0Ijp7InVpZCI6IjI5Nzg0MjM0OTI4OTU0MTAwMDBWMTAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMy42NzIyMTUxMjhaIn0sIm1naWQiOnsidWlkIjoibTU5Z1hyN0xlX1htIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTcuMDk3MDAxNDcxWiJ9LCJuYW5vaW50ZXJhY3RpdmUiOnsidWlkIjoiNmFlYzhjMTAzNzlkY2I3ODQxMmJjODBiNmRkOWM5NzMxNzNhYjdkNzEyZTQzMWE1YTVlYTcwMzRlNTZhNThhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE2LjcxNDgwNzUwNVoifSwib25ldGFnIjp7InVpZCI6IjdPelZoVzFOeC1LOGFVak1HMG52NXVNYm5YNEFHUXZQbnVHcHFrZ3k0ckEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS4xNDE3NDEyNjJaIn0sIm9wZW54Ijp7InVpZCI6IjVkZWNlNjIyLTBhMjMtMGRhYi0zYTI0LTVhNzcwMTBlNDU4MiIsImV4cGlyZXMiOiIyMDIzLTA1LTMxVDA3OjUyOjQ3LjE0MDQxNzM2M1oifSwicHVibWF0aWMiOnsidWlkIjoiN0Q3NUQyNUYtRkFDOS00NDNELUIyRDEtQjE3RkVFMTFFMDI3IiwiZXhwaXJlcyI6IjIwMjItMTAtMzFUMDk6MTQ6MjUuNzM3MjU2ODk5WiJ9LCJyaWNoYXVkaWVuY2UiOnsidWlkIjoiY2I2YzYzMjAtMzNlMi00Nzc0LWIxNjAtMXp6MTY1NDg0MDc0OSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjUyNTA3NDE4WiJ9LCJzbWFydHlhZHMiOnsidWlkIjoiMTJhZjE1ZTQ0ZjAwZDA3NjMwZTc0YzQ5MTU0Y2JmYmE0Zjg0N2U4ZDRhMTU0YzhjM2Q1MWY1OGNmNzJhNDYyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjgyNTAzMTg4NFoifSwic21pbGV3YW50ZWQiOnsidWlkIjoiZGQ5YzNmZTE4N2VmOWIwOWNhYTViNzExNDA0YzI4MzAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNC4yNTU2MDkzNjNaIn0sInN5bmFjb3JtZWRpYSI6eyJ1aWQiOiJHRFBSIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuOTc5NTgzNDM4WiJ9LCJ0cmlwbGVsaWZ0Ijp7InVpZCI6IjcwMjE5NzUwNTQ4MDg4NjUxOTQ2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA4Ljk4OTY3MzU3NFoifSwidmFsdWVpbXByZXNzaW9uIjp7InVpZCI6IjlkMDgxNTVmLWQ5ZmUtNGI1OC04OThlLWUyYzU2MjgyYWIzZSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjA2NzgzOTE2NFoifSwidmlzeCI6eyJ1aWQiOiIyN2UwYWMzYy1iNDZlLTQxYjMtOTkyYy1mOGQyNzE0OTQ5NWUiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi45ODk1MjM1NzNaIn0sInlpZWxkbGFiIjp7InVpZCI6IjY5NzE0ZDlmLWZiMDAtNGE1Zi04MTljLTRiZTE5MTM2YTMyNSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjExLjMwMzAyNjYxNVoifSwieWllbGRtbyI6eyJ1aWQiOiJnOTZjMmY3MTlmMTU1MWIzMWY2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjExMDUyODYwOVoifSwieWllbGRvbmUiOnsidWlkIjoiMmE0MmZiZDMtMmM3MC00ZWI5LWIxYmQtMDQ2OTY2NTBkOTQ4IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuMzE4MzMzOTM5WiJ9LCJ6ZXJvY2xpY2tmcmF1ZCI6eyJ1aWQiOiJiOTk5NThmZS0yYTg3LTJkYTQtOWNjNC05NjFmZDExM2JlY2UiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNS43MTk1OTQ1NjZaIn19LCJiZGF5IjoiMjAyMi0wNS0xN1QwNjo0ODozOC4wMTc5ODgyMDZaIn0=; KRTBCOOKIE_153=1923-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&19420-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&22979-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&23462-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse; KRTBCOOKIE_57=22776-41928985301451193&KRTB&23339-41928985301451193; KRTBCOOKIE_27=16735-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&16736-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&23019-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&23114-uid:3cab6283-4546-4500-a7b6-40ef605fe745; KRTBCOOKIE_18=22947-1978557989514665832; KRTBCOOKIE_466=16530-4fc36250-d852-459c-8772-7356de17ab97; KRTBCOOKIE_391=22924-8044608333778839078&KRTB&23263-8044608333778839078&KRTB&23481-8044608333778839078; KRTBCOOKIE_1310=23431-b81c3g7dr67i&KRTB&23446-b81c3g7dr67i&KRTB&23465-b81c3g7dr67i; KRTBCOOKIE_1290=23368-vkf3yv9lbbl; KRTBCOOKIE_22=14911-4554572065121110164&KRTB&23150-4554572065121110164; KRTBCOOKIE_860=16335-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23334-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23417-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23426-YGAqDU1zUTdjyAFxCoe3kctlNPo; KRTBCOOKIE_904=16787-KwJwE7NkCZClNJRysN2iYg; KRTBCOOKIE_1159=23138-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23328-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23427-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23445-5545f53f3d6e4ec199d8ed627ff026f3; KRTBCOOKIE_32=11175-AQEI_1QecY2ESAIjEW6KAQEBAQE&KRTB&22713-AQEI_1QecY2ESAIjEW6KAQEBAQE&KRTB&22715-AQEI_1QecY2ESAIjEW6KAQEBAQE; SyncRTB3=1685577600%3A35%7C1685491200%3A107_21_71_56_204_247_165_231_233_179_22_209_54_254_238_96_99_220_7_214_13_3_8_234_176_46_5%7C1684886400%3A2_223_15%7C1689465600%3A69%7C1685145600%3A63; KRTBCOOKIE_107=1471-uid:EK38R0PM1NQR0H5&KRTB&23421-uid:EK38R0PM1NQR0H5; KRTBCOOKIE_594=17105-RX-447a6332-530e-456a-97f4-3f0fd1ed48c9-004&KRTB&17107-RX-447a6332-530e-456a-97f4-3f0fd1ed48c9-004; SPugT=1684310122; chkChromeAb67Sec=133; KRTBCOOKIE_699=22727-AAFy2k7FBosAAEasbJoXnw; PugT=1684310473; origin=go-test`)
						return r
					}(),
					Body: []byte(`{"imp":[{"tagid":"/43743431/DMDemo","ext":{"prebid":{"bidder":{"pubmatic":{"publisherId":"5890"}},"adunitcode":"div-gpt-ad-1460505748561-0"}},"id":"div-gpt-ad-1460505748561-0","banner":{"topframe":1,"format":[{"w":300,"h":250}]}}],"site":{"domain":"localhost:9999","publisher":{"domain":"localhost:9999","id":"5890"},"page":"http://localhost:9999/integrationExamples/gpt/owServer_example.html"},"device":{"w":1792,"h":446,"dnt":0,"ua":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36","language":"en","sua":{"source":1,"platform":{"brand":"macOS"},"browsers":[{"brand":"Google Chrome","version":["117"]},{"brand":"Not;A=Brand","version":["8"]},{"brand":"Chromium","version":["117"]}],"mobile":0}},"ext":{"prebid":{"auctiontimestamp":1697191822565,"targeting":{"includewinners":true,"includebidderkeys":true},"bidderparams":{"pubmatic":{"publisherId":"5890","wrapper":{"profileid":43563,"versionid":1}}},"channel":{"name":"pbjs","version":"v8.7.0-pre"},"createtids":false,"debug":true}},"id":"5bdd7da5-1166-40fe-a9cb-3bf3c3164cd3","test":0,"tmax":3000}`),
				},
				setup: func(mme *mock_metrics.MockMetricsEngine) {},
			},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{
				ModuleContext: hookstage.ModuleContext{
					"rctx": models.RequestCtx{
						ProfileID:        43563,
						PubID:            5890,
						PubIDStr:         "5890",
						DisplayID:        1,
						DisplayVersionID: 1,
						SSAuction:        -1,
						Debug:            true,
						DeviceCtx:        models.DeviceCtx{IP: "127.0.0.1", UA: "go-test"},
						IsCTVRequest:     false,
						UidCookie: &http.Cookie{
							Name:  "uids",
							Value: `eyJ0ZW1wVUlEcyI6eyIzM2Fjcm9zcyI6eyJ1aWQiOiIxMTkxNzkxMDk5Nzc2NjEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTo0My4zODg4Nzk5NVoifSwiYWRmIjp7InVpZCI6IjgwNDQ2MDgzMzM3Nzg4MzkwNzgiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMS4wMzMwNTQ3MjdaIn0sImFka2VybmVsIjp7InVpZCI6IkE5MTYzNTAwNzE0OTkyOTMyOTkwIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuMzczMzg1NjYyWiJ9LCJhZGtlcm5lbEFkbiI6eyJ1aWQiOiJBOTE2MzUwMDcxNDk5MjkzMjk5MCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEzLjQzNDkyNTg5NloifSwiYWRtaXhlciI6eyJ1aWQiOiIzNjZhMTdiMTJmMjI0ZDMwOGYzZTNiOGRhOGMzYzhhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjU5MjkxNDgwMVoifSwiYWRueHMiOnsidWlkIjoiNDE5Mjg5ODUzMDE0NTExOTMiLCJleHBpcmVzIjoiMjAyMy0wMS0xOFQwOTo1MzowOC44MjU0NDI2NzZaIn0sImFqYSI6eyJ1aWQiOiJzMnN1aWQ2RGVmMFl0bjJveGQ1aG9zS1AxVmV3IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTMuMjM5MTc2MDU0WiJ9LCJlcGxhbm5pbmciOnsidWlkIjoiQUoxRjBTOE5qdTdTQ0xWOSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjkyOTk2MDQ3M1oifSwiZ2Ftb3NoaSI6eyJ1aWQiOiJndXNyXzM1NmFmOWIxZDhjNjQyYjQ4MmNiYWQyYjdhMjg4MTYxIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuNTI0MTU3MjI1WiJ9LCJncmlkIjp7InVpZCI6IjRmYzM2MjUwLWQ4NTItNDU5Yy04NzcyLTczNTZkZTE3YWI5NyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE0LjY5NjMxNjIyN1oifSwiZ3JvdXBtIjp7InVpZCI6IjdENzVEMjVGLUZBQzktNDQzRC1CMkQxLUIxN0ZFRTExRTAyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjM5LjIyNjIxMjUzMloifSwiaXgiOnsidWlkIjoiWW9ORlNENlc5QkphOEh6eEdtcXlCUUFBXHUwMDI2Mjk3IiwiZXhwaXJlcyI6IjIwMjMtMDUtMzFUMDc6NTM6MzguNTU1ODI3MzU0WiJ9LCJqaXhpZSI6eyJ1aWQiOiI3MzY3MTI1MC1lODgyLTExZWMtYjUzOC0xM2FjYjdhZjBkZTQiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi4xOTEwOTk3MzJaIn0sImxvZ2ljYWQiOnsidWlkIjoiQVZ4OVROQS11c25pa3M4QURzTHpWa3JvaDg4QUFBR0JUREh0UUEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS40NTUxNDk2MTZaIn0sIm1lZGlhbmV0Ijp7InVpZCI6IjI5Nzg0MjM0OTI4OTU0MTAwMDBWMTAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMy42NzIyMTUxMjhaIn0sIm1naWQiOnsidWlkIjoibTU5Z1hyN0xlX1htIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTcuMDk3MDAxNDcxWiJ9LCJuYW5vaW50ZXJhY3RpdmUiOnsidWlkIjoiNmFlYzhjMTAzNzlkY2I3ODQxMmJjODBiNmRkOWM5NzMxNzNhYjdkNzEyZTQzMWE1YTVlYTcwMzRlNTZhNThhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE2LjcxNDgwNzUwNVoifSwib25ldGFnIjp7InVpZCI6IjdPelZoVzFOeC1LOGFVak1HMG52NXVNYm5YNEFHUXZQbnVHcHFrZ3k0ckEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS4xNDE3NDEyNjJaIn0sIm9wZW54Ijp7InVpZCI6IjVkZWNlNjIyLTBhMjMtMGRhYi0zYTI0LTVhNzcwMTBlNDU4MiIsImV4cGlyZXMiOiIyMDIzLTA1LTMxVDA3OjUyOjQ3LjE0MDQxNzM2M1oifSwicHVibWF0aWMiOnsidWlkIjoiN0Q3NUQyNUYtRkFDOS00NDNELUIyRDEtQjE3RkVFMTFFMDI3IiwiZXhwaXJlcyI6IjIwMjItMTAtMzFUMDk6MTQ6MjUuNzM3MjU2ODk5WiJ9LCJyaWNoYXVkaWVuY2UiOnsidWlkIjoiY2I2YzYzMjAtMzNlMi00Nzc0LWIxNjAtMXp6MTY1NDg0MDc0OSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjUyNTA3NDE4WiJ9LCJzbWFydHlhZHMiOnsidWlkIjoiMTJhZjE1ZTQ0ZjAwZDA3NjMwZTc0YzQ5MTU0Y2JmYmE0Zjg0N2U4ZDRhMTU0YzhjM2Q1MWY1OGNmNzJhNDYyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjgyNTAzMTg4NFoifSwic21pbGV3YW50ZWQiOnsidWlkIjoiZGQ5YzNmZTE4N2VmOWIwOWNhYTViNzExNDA0YzI4MzAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNC4yNTU2MDkzNjNaIn0sInN5bmFjb3JtZWRpYSI6eyJ1aWQiOiJHRFBSIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuOTc5NTgzNDM4WiJ9LCJ0cmlwbGVsaWZ0Ijp7InVpZCI6IjcwMjE5NzUwNTQ4MDg4NjUxOTQ2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA4Ljk4OTY3MzU3NFoifSwidmFsdWVpbXByZXNzaW9uIjp7InVpZCI6IjlkMDgxNTVmLWQ5ZmUtNGI1OC04OThlLWUyYzU2MjgyYWIzZSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjA2NzgzOTE2NFoifSwidmlzeCI6eyJ1aWQiOiIyN2UwYWMzYy1iNDZlLTQxYjMtOTkyYy1mOGQyNzE0OTQ5NWUiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi45ODk1MjM1NzNaIn0sInlpZWxkbGFiIjp7InVpZCI6IjY5NzE0ZDlmLWZiMDAtNGE1Zi04MTljLTRiZTE5MTM2YTMyNSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjExLjMwMzAyNjYxNVoifSwieWllbGRtbyI6eyJ1aWQiOiJnOTZjMmY3MTlmMTU1MWIzMWY2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjExMDUyODYwOVoifSwieWllbGRvbmUiOnsidWlkIjoiMmE0MmZiZDMtMmM3MC00ZWI5LWIxYmQtMDQ2OTY2NTBkOTQ4IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuMzE4MzMzOTM5WiJ9LCJ6ZXJvY2xpY2tmcmF1ZCI6eyJ1aWQiOiJiOTk5NThmZS0yYTg3LTJkYTQtOWNjNC05NjFmZDExM2JlY2UiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNS43MTk1OTQ1NjZaIn19LCJiZGF5IjoiMjAyMi0wNS0xN1QwNjo0ODozOC4wMTc5ODgyMDZaIn0=`,
						},
						KADUSERCookie: &http.Cookie{
							Name:  "KADUSERCOOKIE",
							Value: `7D75D25F-FAC9-443D-B2D1-B17FEE11E027`,
						},
						OriginCookie:                    "go-test",
						Aliases:                         make(map[string]string),
						ImpBidCtx:                       make(map[string]models.ImpCtx),
						PrebidBidderCode:                make(map[string]string),
						BidderResponseTimeMillis:        make(map[string]int),
						ProfileIDStr:                    "43563",
						Endpoint:                        models.EndpointWebS2S,
						MetricsEngine:                   mockEngine,
						SeatNonBids:                     make(map[string][]openrtb_ext.NonBid),
						Method:                          "POST",
						WakandaDebug:                    &wakanda.Debug{},
						ImpCountingMethodEnabledBidders: make(map[string]struct{}),
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "/openrtb2/auction?source=pbjs request(webs2s) without profileid",
			fields: fields{
				cfg:   config.Config{},
				cache: nil,
			},
			args: args{
				in0:   context.Background(),
				miCtx: hookstage.ModuleInvocationContext{},
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						r, err := http.NewRequest("POST", "http://localhost/openrtb2/auction?source=pbjs&debug=1", nil)
						if err != nil {
							panic(err)
						}
						r.Header.Add("User-Agent", "go-test")
						r.Header.Add("SOURCE_IP", "127.0.0.1")
						r.Header.Add("Cookie", `KADUSERCOOKIE=7D75D25F-FAC9-443D-B2D1-B17FEE11E027; DPSync3=1684886400%3A248%7C1685491200%3A245_226_201; KRTBCOOKIE_80=16514-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&22987-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&23025-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&23386-CAESEMih0bN7ISRdZT8xX8LXzEw; KRTBCOOKIE_377=6810-59dc50c9-d658-44ce-b442-5a1f344d97c0&KRTB&22918-59dc50c9-d658-44ce-b442-5a1f344d97c0&KRTB&23031-59dc50c9-d658-44ce-b442-5a1f344d97c0; uids=eyJ0ZW1wVUlEcyI6eyIzM2Fjcm9zcyI6eyJ1aWQiOiIxMTkxNzkxMDk5Nzc2NjEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTo0My4zODg4Nzk5NVoifSwiYWRmIjp7InVpZCI6IjgwNDQ2MDgzMzM3Nzg4MzkwNzgiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMS4wMzMwNTQ3MjdaIn0sImFka2VybmVsIjp7InVpZCI6IkE5MTYzNTAwNzE0OTkyOTMyOTkwIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuMzczMzg1NjYyWiJ9LCJhZGtlcm5lbEFkbiI6eyJ1aWQiOiJBOTE2MzUwMDcxNDk5MjkzMjk5MCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEzLjQzNDkyNTg5NloifSwiYWRtaXhlciI6eyJ1aWQiOiIzNjZhMTdiMTJmMjI0ZDMwOGYzZTNiOGRhOGMzYzhhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjU5MjkxNDgwMVoifSwiYWRueHMiOnsidWlkIjoiNDE5Mjg5ODUzMDE0NTExOTMiLCJleHBpcmVzIjoiMjAyMy0wMS0xOFQwOTo1MzowOC44MjU0NDI2NzZaIn0sImFqYSI6eyJ1aWQiOiJzMnN1aWQ2RGVmMFl0bjJveGQ1aG9zS1AxVmV3IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTMuMjM5MTc2MDU0WiJ9LCJlcGxhbm5pbmciOnsidWlkIjoiQUoxRjBTOE5qdTdTQ0xWOSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjkyOTk2MDQ3M1oifSwiZ2Ftb3NoaSI6eyJ1aWQiOiJndXNyXzM1NmFmOWIxZDhjNjQyYjQ4MmNiYWQyYjdhMjg4MTYxIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuNTI0MTU3MjI1WiJ9LCJncmlkIjp7InVpZCI6IjRmYzM2MjUwLWQ4NTItNDU5Yy04NzcyLTczNTZkZTE3YWI5NyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE0LjY5NjMxNjIyN1oifSwiZ3JvdXBtIjp7InVpZCI6IjdENzVEMjVGLUZBQzktNDQzRC1CMkQxLUIxN0ZFRTExRTAyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjM5LjIyNjIxMjUzMloifSwiaXgiOnsidWlkIjoiWW9ORlNENlc5QkphOEh6eEdtcXlCUUFBXHUwMDI2Mjk3IiwiZXhwaXJlcyI6IjIwMjMtMDUtMzFUMDc6NTM6MzguNTU1ODI3MzU0WiJ9LCJqaXhpZSI6eyJ1aWQiOiI3MzY3MTI1MC1lODgyLTExZWMtYjUzOC0xM2FjYjdhZjBkZTQiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi4xOTEwOTk3MzJaIn0sImxvZ2ljYWQiOnsidWlkIjoiQVZ4OVROQS11c25pa3M4QURzTHpWa3JvaDg4QUFBR0JUREh0UUEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS40NTUxNDk2MTZaIn0sIm1lZGlhbmV0Ijp7InVpZCI6IjI5Nzg0MjM0OTI4OTU0MTAwMDBWMTAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMy42NzIyMTUxMjhaIn0sIm1naWQiOnsidWlkIjoibTU5Z1hyN0xlX1htIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTcuMDk3MDAxNDcxWiJ9LCJuYW5vaW50ZXJhY3RpdmUiOnsidWlkIjoiNmFlYzhjMTAzNzlkY2I3ODQxMmJjODBiNmRkOWM5NzMxNzNhYjdkNzEyZTQzMWE1YTVlYTcwMzRlNTZhNThhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE2LjcxNDgwNzUwNVoifSwib25ldGFnIjp7InVpZCI6IjdPelZoVzFOeC1LOGFVak1HMG52NXVNYm5YNEFHUXZQbnVHcHFrZ3k0ckEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS4xNDE3NDEyNjJaIn0sIm9wZW54Ijp7InVpZCI6IjVkZWNlNjIyLTBhMjMtMGRhYi0zYTI0LTVhNzcwMTBlNDU4MiIsImV4cGlyZXMiOiIyMDIzLTA1LTMxVDA3OjUyOjQ3LjE0MDQxNzM2M1oifSwicHVibWF0aWMiOnsidWlkIjoiN0Q3NUQyNUYtRkFDOS00NDNELUIyRDEtQjE3RkVFMTFFMDI3IiwiZXhwaXJlcyI6IjIwMjItMTAtMzFUMDk6MTQ6MjUuNzM3MjU2ODk5WiJ9LCJyaWNoYXVkaWVuY2UiOnsidWlkIjoiY2I2YzYzMjAtMzNlMi00Nzc0LWIxNjAtMXp6MTY1NDg0MDc0OSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjUyNTA3NDE4WiJ9LCJzbWFydHlhZHMiOnsidWlkIjoiMTJhZjE1ZTQ0ZjAwZDA3NjMwZTc0YzQ5MTU0Y2JmYmE0Zjg0N2U4ZDRhMTU0YzhjM2Q1MWY1OGNmNzJhNDYyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjgyNTAzMTg4NFoifSwic21pbGV3YW50ZWQiOnsidWlkIjoiZGQ5YzNmZTE4N2VmOWIwOWNhYTViNzExNDA0YzI4MzAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNC4yNTU2MDkzNjNaIn0sInN5bmFjb3JtZWRpYSI6eyJ1aWQiOiJHRFBSIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuOTc5NTgzNDM4WiJ9LCJ0cmlwbGVsaWZ0Ijp7InVpZCI6IjcwMjE5NzUwNTQ4MDg4NjUxOTQ2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA4Ljk4OTY3MzU3NFoifSwidmFsdWVpbXByZXNzaW9uIjp7InVpZCI6IjlkMDgxNTVmLWQ5ZmUtNGI1OC04OThlLWUyYzU2MjgyYWIzZSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjA2NzgzOTE2NFoifSwidmlzeCI6eyJ1aWQiOiIyN2UwYWMzYy1iNDZlLTQxYjMtOTkyYy1mOGQyNzE0OTQ5NWUiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi45ODk1MjM1NzNaIn0sInlpZWxkbGFiIjp7InVpZCI6IjY5NzE0ZDlmLWZiMDAtNGE1Zi04MTljLTRiZTE5MTM2YTMyNSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjExLjMwMzAyNjYxNVoifSwieWllbGRtbyI6eyJ1aWQiOiJnOTZjMmY3MTlmMTU1MWIzMWY2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjExMDUyODYwOVoifSwieWllbGRvbmUiOnsidWlkIjoiMmE0MmZiZDMtMmM3MC00ZWI5LWIxYmQtMDQ2OTY2NTBkOTQ4IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuMzE4MzMzOTM5WiJ9LCJ6ZXJvY2xpY2tmcmF1ZCI6eyJ1aWQiOiJiOTk5NThmZS0yYTg3LTJkYTQtOWNjNC05NjFmZDExM2JlY2UiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNS43MTk1OTQ1NjZaIn19LCJiZGF5IjoiMjAyMi0wNS0xN1QwNjo0ODozOC4wMTc5ODgyMDZaIn0=; KRTBCOOKIE_153=1923-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&19420-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&22979-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&23462-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse; KRTBCOOKIE_57=22776-41928985301451193&KRTB&23339-41928985301451193; KRTBCOOKIE_27=16735-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&16736-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&23019-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&23114-uid:3cab6283-4546-4500-a7b6-40ef605fe745; KRTBCOOKIE_18=22947-1978557989514665832; KRTBCOOKIE_466=16530-4fc36250-d852-459c-8772-7356de17ab97; KRTBCOOKIE_391=22924-8044608333778839078&KRTB&23263-8044608333778839078&KRTB&23481-8044608333778839078; KRTBCOOKIE_1310=23431-b81c3g7dr67i&KRTB&23446-b81c3g7dr67i&KRTB&23465-b81c3g7dr67i; KRTBCOOKIE_1290=23368-vkf3yv9lbbl; KRTBCOOKIE_22=14911-4554572065121110164&KRTB&23150-4554572065121110164; KRTBCOOKIE_860=16335-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23334-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23417-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23426-YGAqDU1zUTdjyAFxCoe3kctlNPo; KRTBCOOKIE_904=16787-KwJwE7NkCZClNJRysN2iYg; KRTBCOOKIE_1159=23138-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23328-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23427-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23445-5545f53f3d6e4ec199d8ed627ff026f3; KRTBCOOKIE_32=11175-AQEI_1QecY2ESAIjEW6KAQEBAQE&KRTB&22713-AQEI_1QecY2ESAIjEW6KAQEBAQE&KRTB&22715-AQEI_1QecY2ESAIjEW6KAQEBAQE; SyncRTB3=1685577600%3A35%7C1685491200%3A107_21_71_56_204_247_165_231_233_179_22_209_54_254_238_96_99_220_7_214_13_3_8_234_176_46_5%7C1684886400%3A2_223_15%7C1689465600%3A69%7C1685145600%3A63; KRTBCOOKIE_107=1471-uid:EK38R0PM1NQR0H5&KRTB&23421-uid:EK38R0PM1NQR0H5; KRTBCOOKIE_594=17105-RX-447a6332-530e-456a-97f4-3f0fd1ed48c9-004&KRTB&17107-RX-447a6332-530e-456a-97f4-3f0fd1ed48c9-004; SPugT=1684310122; chkChromeAb67Sec=133; KRTBCOOKIE_699=22727-AAFy2k7FBosAAEasbJoXnw; PugT=1684310473; origin=go-test`)
						return r
					}(),
					Body: []byte(`{"imp":[{"tagid":"/43743431/DMDemo","ext":{"prebid":{"bidder":{"pubmatic":{"publisherId":"5890"}},"adunitcode":"div-gpt-ad-1460505748561-0"}},"id":"div-gpt-ad-1460505748561-0","banner":{"topframe":1,"format":[{"w":300,"h":250}]}}],"site":{"domain":"localhost:9999","publisher":{"domain":"localhost:9999","id":"5890"},"page":"http://localhost:9999/integrationExamples/gpt/owServer_example.html"},"device":{"w":1792,"h":446,"dnt":0,"ua":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36","language":"en","sua":{"source":1,"platform":{"brand":"macOS"},"browsers":[{"brand":"Google Chrome","version":["117"]},{"brand":"Not;A=Brand","version":["8"]},{"brand":"Chromium","version":["117"]}],"mobile":0}},"ext":{"prebid":{"auctiontimestamp":1697191822565,"targeting":{"includewinners":true,"includebidderkeys":true},"bidderparams":{"pubmatic":{"publisherId":"5890","wrapper":{}}},"channel":{"name":"pbjs","version":"v8.7.0-pre"},"createtids":false}},"id":"5bdd7da5-1166-40fe-a9cb-3bf3c3164cd3","test":0,"tmax":3000}`),
				},
				setup: func(mme *mock_metrics.MockMetricsEngine) {
					mme.EXPECT().RecordBadRequests(gomock.Any(), "0", 700)
				},
			},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{
				Reject:  true,
				NbrCode: int(nbr.InvalidProfileID),
				Errors:  []string{"ErrMissingProfileID"},
			},
			wantErr: nil,
		},
		{
			name: "/openrtb2/auction with source!=(pbjs or owsdk). new hybrid endpoint should not execute module",
			fields: fields{
				cfg:   config.Config{},
				cache: nil,
			},
			args: args{
				in0:   context.Background(),
				miCtx: hookstage.ModuleInvocationContext{},
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						r, err := http.NewRequest("POST", "http://localhost/openrtb2/auction?debug=1", nil)
						if err != nil {
							panic(err)
						}
						r.Header.Add("User-Agent", "go-test")
						r.Header.Add("SOURCE_IP", "127.0.0.1")
						r.Header.Add("Cookie", `KADUSERCOOKIE=7D75D25F-FAC9-443D-B2D1-B17FEE11E027; DPSync3=1684886400%3A248%7C1685491200%3A245_226_201; KRTBCOOKIE_80=16514-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&22987-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&23025-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&23386-CAESEMih0bN7ISRdZT8xX8LXzEw; KRTBCOOKIE_377=6810-59dc50c9-d658-44ce-b442-5a1f344d97c0&KRTB&22918-59dc50c9-d658-44ce-b442-5a1f344d97c0&KRTB&23031-59dc50c9-d658-44ce-b442-5a1f344d97c0; uids=eyJ0ZW1wVUlEcyI6eyIzM2Fjcm9zcyI6eyJ1aWQiOiIxMTkxNzkxMDk5Nzc2NjEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTo0My4zODg4Nzk5NVoifSwiYWRmIjp7InVpZCI6IjgwNDQ2MDgzMzM3Nzg4MzkwNzgiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMS4wMzMwNTQ3MjdaIn0sImFka2VybmVsIjp7InVpZCI6IkE5MTYzNTAwNzE0OTkyOTMyOTkwIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuMzczMzg1NjYyWiJ9LCJhZGtlcm5lbEFkbiI6eyJ1aWQiOiJBOTE2MzUwMDcxNDk5MjkzMjk5MCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEzLjQzNDkyNTg5NloifSwiYWRtaXhlciI6eyJ1aWQiOiIzNjZhMTdiMTJmMjI0ZDMwOGYzZTNiOGRhOGMzYzhhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjU5MjkxNDgwMVoifSwiYWRueHMiOnsidWlkIjoiNDE5Mjg5ODUzMDE0NTExOTMiLCJleHBpcmVzIjoiMjAyMy0wMS0xOFQwOTo1MzowOC44MjU0NDI2NzZaIn0sImFqYSI6eyJ1aWQiOiJzMnN1aWQ2RGVmMFl0bjJveGQ1aG9zS1AxVmV3IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTMuMjM5MTc2MDU0WiJ9LCJlcGxhbm5pbmciOnsidWlkIjoiQUoxRjBTOE5qdTdTQ0xWOSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjkyOTk2MDQ3M1oifSwiZ2Ftb3NoaSI6eyJ1aWQiOiJndXNyXzM1NmFmOWIxZDhjNjQyYjQ4MmNiYWQyYjdhMjg4MTYxIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuNTI0MTU3MjI1WiJ9LCJncmlkIjp7InVpZCI6IjRmYzM2MjUwLWQ4NTItNDU5Yy04NzcyLTczNTZkZTE3YWI5NyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE0LjY5NjMxNjIyN1oifSwiZ3JvdXBtIjp7InVpZCI6IjdENzVEMjVGLUZBQzktNDQzRC1CMkQxLUIxN0ZFRTExRTAyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjM5LjIyNjIxMjUzMloifSwiaXgiOnsidWlkIjoiWW9ORlNENlc5QkphOEh6eEdtcXlCUUFBXHUwMDI2Mjk3IiwiZXhwaXJlcyI6IjIwMjMtMDUtMzFUMDc6NTM6MzguNTU1ODI3MzU0WiJ9LCJqaXhpZSI6eyJ1aWQiOiI3MzY3MTI1MC1lODgyLTExZWMtYjUzOC0xM2FjYjdhZjBkZTQiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi4xOTEwOTk3MzJaIn0sImxvZ2ljYWQiOnsidWlkIjoiQVZ4OVROQS11c25pa3M4QURzTHpWa3JvaDg4QUFBR0JUREh0UUEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS40NTUxNDk2MTZaIn0sIm1lZGlhbmV0Ijp7InVpZCI6IjI5Nzg0MjM0OTI4OTU0MTAwMDBWMTAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMy42NzIyMTUxMjhaIn0sIm1naWQiOnsidWlkIjoibTU5Z1hyN0xlX1htIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTcuMDk3MDAxNDcxWiJ9LCJuYW5vaW50ZXJhY3RpdmUiOnsidWlkIjoiNmFlYzhjMTAzNzlkY2I3ODQxMmJjODBiNmRkOWM5NzMxNzNhYjdkNzEyZTQzMWE1YTVlYTcwMzRlNTZhNThhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE2LjcxNDgwNzUwNVoifSwib25ldGFnIjp7InVpZCI6IjdPelZoVzFOeC1LOGFVak1HMG52NXVNYm5YNEFHUXZQbnVHcHFrZ3k0ckEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS4xNDE3NDEyNjJaIn0sIm9wZW54Ijp7InVpZCI6IjVkZWNlNjIyLTBhMjMtMGRhYi0zYTI0LTVhNzcwMTBlNDU4MiIsImV4cGlyZXMiOiIyMDIzLTA1LTMxVDA3OjUyOjQ3LjE0MDQxNzM2M1oifSwicHVibWF0aWMiOnsidWlkIjoiN0Q3NUQyNUYtRkFDOS00NDNELUIyRDEtQjE3RkVFMTFFMDI3IiwiZXhwaXJlcyI6IjIwMjItMTAtMzFUMDk6MTQ6MjUuNzM3MjU2ODk5WiJ9LCJyaWNoYXVkaWVuY2UiOnsidWlkIjoiY2I2YzYzMjAtMzNlMi00Nzc0LWIxNjAtMXp6MTY1NDg0MDc0OSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjUyNTA3NDE4WiJ9LCJzbWFydHlhZHMiOnsidWlkIjoiMTJhZjE1ZTQ0ZjAwZDA3NjMwZTc0YzQ5MTU0Y2JmYmE0Zjg0N2U4ZDRhMTU0YzhjM2Q1MWY1OGNmNzJhNDYyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjgyNTAzMTg4NFoifSwic21pbGV3YW50ZWQiOnsidWlkIjoiZGQ5YzNmZTE4N2VmOWIwOWNhYTViNzExNDA0YzI4MzAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNC4yNTU2MDkzNjNaIn0sInN5bmFjb3JtZWRpYSI6eyJ1aWQiOiJHRFBSIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuOTc5NTgzNDM4WiJ9LCJ0cmlwbGVsaWZ0Ijp7InVpZCI6IjcwMjE5NzUwNTQ4MDg4NjUxOTQ2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA4Ljk4OTY3MzU3NFoifSwidmFsdWVpbXByZXNzaW9uIjp7InVpZCI6IjlkMDgxNTVmLWQ5ZmUtNGI1OC04OThlLWUyYzU2MjgyYWIzZSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjA2NzgzOTE2NFoifSwidmlzeCI6eyJ1aWQiOiIyN2UwYWMzYy1iNDZlLTQxYjMtOTkyYy1mOGQyNzE0OTQ5NWUiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi45ODk1MjM1NzNaIn0sInlpZWxkbGFiIjp7InVpZCI6IjY5NzE0ZDlmLWZiMDAtNGE1Zi04MTljLTRiZTE5MTM2YTMyNSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjExLjMwMzAyNjYxNVoifSwieWllbGRtbyI6eyJ1aWQiOiJnOTZjMmY3MTlmMTU1MWIzMWY2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjExMDUyODYwOVoifSwieWllbGRvbmUiOnsidWlkIjoiMmE0MmZiZDMtMmM3MC00ZWI5LWIxYmQtMDQ2OTY2NTBkOTQ4IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuMzE4MzMzOTM5WiJ9LCJ6ZXJvY2xpY2tmcmF1ZCI6eyJ1aWQiOiJiOTk5NThmZS0yYTg3LTJkYTQtOWNjNC05NjFmZDExM2JlY2UiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNS43MTk1OTQ1NjZaIn19LCJiZGF5IjoiMjAyMi0wNS0xN1QwNjo0ODozOC4wMTc5ODgyMDZaIn0=; KRTBCOOKIE_153=1923-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&19420-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&22979-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&23462-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse; KRTBCOOKIE_57=22776-41928985301451193&KRTB&23339-41928985301451193; KRTBCOOKIE_27=16735-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&16736-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&23019-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&23114-uid:3cab6283-4546-4500-a7b6-40ef605fe745; KRTBCOOKIE_18=22947-1978557989514665832; KRTBCOOKIE_466=16530-4fc36250-d852-459c-8772-7356de17ab97; KRTBCOOKIE_391=22924-8044608333778839078&KRTB&23263-8044608333778839078&KRTB&23481-8044608333778839078; KRTBCOOKIE_1310=23431-b81c3g7dr67i&KRTB&23446-b81c3g7dr67i&KRTB&23465-b81c3g7dr67i; KRTBCOOKIE_1290=23368-vkf3yv9lbbl; KRTBCOOKIE_22=14911-4554572065121110164&KRTB&23150-4554572065121110164; KRTBCOOKIE_860=16335-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23334-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23417-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23426-YGAqDU1zUTdjyAFxCoe3kctlNPo; KRTBCOOKIE_904=16787-KwJwE7NkCZClNJRysN2iYg; KRTBCOOKIE_1159=23138-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23328-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23427-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23445-5545f53f3d6e4ec199d8ed627ff026f3; KRTBCOOKIE_32=11175-AQEI_1QecY2ESAIjEW6KAQEBAQE&KRTB&22713-AQEI_1QecY2ESAIjEW6KAQEBAQE&KRTB&22715-AQEI_1QecY2ESAIjEW6KAQEBAQE; SyncRTB3=1685577600%3A35%7C1685491200%3A107_21_71_56_204_247_165_231_233_179_22_209_54_254_238_96_99_220_7_214_13_3_8_234_176_46_5%7C1684886400%3A2_223_15%7C1689465600%3A69%7C1685145600%3A63; KRTBCOOKIE_107=1471-uid:EK38R0PM1NQR0H5&KRTB&23421-uid:EK38R0PM1NQR0H5; KRTBCOOKIE_594=17105-RX-447a6332-530e-456a-97f4-3f0fd1ed48c9-004&KRTB&17107-RX-447a6332-530e-456a-97f4-3f0fd1ed48c9-004; SPugT=1684310122; chkChromeAb67Sec=133; KRTBCOOKIE_699=22727-AAFy2k7FBosAAEasbJoXnw; PugT=1684310473; origin=go-test`)
						return r
					}(),
					Body: []byte(`{"imp":[{"tagid":"/43743431/DMDemo","ext":{"prebid":{"bidder":{"pubmatic":{"publisherId":"5890"}},"adunitcode":"div-gpt-ad-1460505748561-0"}},"id":"div-gpt-ad-1460505748561-0","banner":{"topframe":1,"format":[{"w":300,"h":250}]}}],"site":{"domain":"localhost:9999","publisher":{"domain":"localhost:9999","id":"5890"},"page":"http://localhost:9999/integrationExamples/gpt/owServer_example.html"},"device":{"w":1792,"h":446,"dnt":0,"ua":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36","language":"en","sua":{"source":1,"platform":{"brand":"macOS"},"browsers":[{"brand":"Google Chrome","version":["117"]},{"brand":"Not;A=Brand","version":["8"]},{"brand":"Chromium","version":["117"]}],"mobile":0}},"ext":{"prebid":{"auctiontimestamp":1697191822565,"targeting":{"includewinners":true,"includebidderkeys":true},"bidderparams":{"pubmatic":{"publisherId":"5890","wrapper":{}}},"channel":{"name":"pbjs","version":"v8.7.0-pre"},"createtids":false}},"id":"5bdd7da5-1166-40fe-a9cb-3bf3c3164cd3","test":0,"tmax":3000}`),
				},
				setup: func(mme *mock_metrics.MockMetricsEngine) {},
			},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{
				ModuleContext: hookstage.ModuleContext{
					"rctx": models.RequestCtx{
						Endpoint: models.EndpointHybrid,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "hybrid api pbs/openrtb2/auction should not execute entrypointhook",
			args: args{
				in0:   context.Background(),
				miCtx: hookstage.ModuleInvocationContext{},
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						r, err := http.NewRequest("POST", "http://localhost/pbs/openrtb2/auction", nil)
						if err != nil {
							panic(err)
						}
						r.Header.Add("User-Agent", "go-test")
						r.Header.Add("SOURCE_IP", "127.0.0.1")
						r.Header.Add("Cookie", `KADUSERCOOKIE=7D75D25F-FAC9-443D-B2D1-B17FEE11E027; DPSync3=1684886400%3A248%7C1685491200%3A245_226_201; KRTBCOOKIE_80=16514-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&22987-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&23025-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&23386-CAESEMih0bN7ISRdZT8xX8LXzEw; KRTBCOOKIE_377=6810-59dc50c9-d658-44ce-b442-5a1f344d97c0&KRTB&22918-59dc50c9-d658-44ce-b442-5a1f344d97c0&KRTB&23031-59dc50c9-d658-44ce-b442-5a1f344d97c0; uids=eyJ0ZW1wVUlEcyI6eyIzM2Fjcm9zcyI6eyJ1aWQiOiIxMTkxNzkxMDk5Nzc2NjEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTo0My4zODg4Nzk5NVoifSwiYWRmIjp7InVpZCI6IjgwNDQ2MDgzMzM3Nzg4MzkwNzgiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMS4wMzMwNTQ3MjdaIn0sImFka2VybmVsIjp7InVpZCI6IkE5MTYzNTAwNzE0OTkyOTMyOTkwIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuMzczMzg1NjYyWiJ9LCJhZGtlcm5lbEFkbiI6eyJ1aWQiOiJBOTE2MzUwMDcxNDk5MjkzMjk5MCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEzLjQzNDkyNTg5NloifSwiYWRtaXhlciI6eyJ1aWQiOiIzNjZhMTdiMTJmMjI0ZDMwOGYzZTNiOGRhOGMzYzhhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjU5MjkxNDgwMVoifSwiYWRueHMiOnsidWlkIjoiNDE5Mjg5ODUzMDE0NTExOTMiLCJleHBpcmVzIjoiMjAyMy0wMS0xOFQwOTo1MzowOC44MjU0NDI2NzZaIn0sImFqYSI6eyJ1aWQiOiJzMnN1aWQ2RGVmMFl0bjJveGQ1aG9zS1AxVmV3IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTMuMjM5MTc2MDU0WiJ9LCJlcGxhbm5pbmciOnsidWlkIjoiQUoxRjBTOE5qdTdTQ0xWOSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjkyOTk2MDQ3M1oifSwiZ2Ftb3NoaSI6eyJ1aWQiOiJndXNyXzM1NmFmOWIxZDhjNjQyYjQ4MmNiYWQyYjdhMjg4MTYxIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuNTI0MTU3MjI1WiJ9LCJncmlkIjp7InVpZCI6IjRmYzM2MjUwLWQ4NTItNDU5Yy04NzcyLTczNTZkZTE3YWI5NyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE0LjY5NjMxNjIyN1oifSwiZ3JvdXBtIjp7InVpZCI6IjdENzVEMjVGLUZBQzktNDQzRC1CMkQxLUIxN0ZFRTExRTAyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjM5LjIyNjIxMjUzMloifSwiaXgiOnsidWlkIjoiWW9ORlNENlc5QkphOEh6eEdtcXlCUUFBXHUwMDI2Mjk3IiwiZXhwaXJlcyI6IjIwMjMtMDUtMzFUMDc6NTM6MzguNTU1ODI3MzU0WiJ9LCJqaXhpZSI6eyJ1aWQiOiI3MzY3MTI1MC1lODgyLTExZWMtYjUzOC0xM2FjYjdhZjBkZTQiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi4xOTEwOTk3MzJaIn0sImxvZ2ljYWQiOnsidWlkIjoiQVZ4OVROQS11c25pa3M4QURzTHpWa3JvaDg4QUFBR0JUREh0UUEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS40NTUxNDk2MTZaIn0sIm1lZGlhbmV0Ijp7InVpZCI6IjI5Nzg0MjM0OTI4OTU0MTAwMDBWMTAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMy42NzIyMTUxMjhaIn0sIm1naWQiOnsidWlkIjoibTU5Z1hyN0xlX1htIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTcuMDk3MDAxNDcxWiJ9LCJuYW5vaW50ZXJhY3RpdmUiOnsidWlkIjoiNmFlYzhjMTAzNzlkY2I3ODQxMmJjODBiNmRkOWM5NzMxNzNhYjdkNzEyZTQzMWE1YTVlYTcwMzRlNTZhNThhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE2LjcxNDgwNzUwNVoifSwib25ldGFnIjp7InVpZCI6IjdPelZoVzFOeC1LOGFVak1HMG52NXVNYm5YNEFHUXZQbnVHcHFrZ3k0ckEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS4xNDE3NDEyNjJaIn0sIm9wZW54Ijp7InVpZCI6IjVkZWNlNjIyLTBhMjMtMGRhYi0zYTI0LTVhNzcwMTBlNDU4MiIsImV4cGlyZXMiOiIyMDIzLTA1LTMxVDA3OjUyOjQ3LjE0MDQxNzM2M1oifSwicHVibWF0aWMiOnsidWlkIjoiN0Q3NUQyNUYtRkFDOS00NDNELUIyRDEtQjE3RkVFMTFFMDI3IiwiZXhwaXJlcyI6IjIwMjItMTAtMzFUMDk6MTQ6MjUuNzM3MjU2ODk5WiJ9LCJyaWNoYXVkaWVuY2UiOnsidWlkIjoiY2I2YzYzMjAtMzNlMi00Nzc0LWIxNjAtMXp6MTY1NDg0MDc0OSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjUyNTA3NDE4WiJ9LCJzbWFydHlhZHMiOnsidWlkIjoiMTJhZjE1ZTQ0ZjAwZDA3NjMwZTc0YzQ5MTU0Y2JmYmE0Zjg0N2U4ZDRhMTU0YzhjM2Q1MWY1OGNmNzJhNDYyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjgyNTAzMTg4NFoifSwic21pbGV3YW50ZWQiOnsidWlkIjoiZGQ5YzNmZTE4N2VmOWIwOWNhYTViNzExNDA0YzI4MzAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNC4yNTU2MDkzNjNaIn0sInN5bmFjb3JtZWRpYSI6eyJ1aWQiOiJHRFBSIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuOTc5NTgzNDM4WiJ9LCJ0cmlwbGVsaWZ0Ijp7InVpZCI6IjcwMjE5NzUwNTQ4MDg4NjUxOTQ2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA4Ljk4OTY3MzU3NFoifSwidmFsdWVpbXByZXNzaW9uIjp7InVpZCI6IjlkMDgxNTVmLWQ5ZmUtNGI1OC04OThlLWUyYzU2MjgyYWIzZSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjA2NzgzOTE2NFoifSwidmlzeCI6eyJ1aWQiOiIyN2UwYWMzYy1iNDZlLTQxYjMtOTkyYy1mOGQyNzE0OTQ5NWUiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi45ODk1MjM1NzNaIn0sInlpZWxkbGFiIjp7InVpZCI6IjY5NzE0ZDlmLWZiMDAtNGE1Zi04MTljLTRiZTE5MTM2YTMyNSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjExLjMwMzAyNjYxNVoifSwieWllbGRtbyI6eyJ1aWQiOiJnOTZjMmY3MTlmMTU1MWIzMWY2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjExMDUyODYwOVoifSwieWllbGRvbmUiOnsidWlkIjoiMmE0MmZiZDMtMmM3MC00ZWI5LWIxYmQtMDQ2OTY2NTBkOTQ4IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuMzE4MzMzOTM5WiJ9LCJ6ZXJvY2xpY2tmcmF1ZCI6eyJ1aWQiOiJiOTk5NThmZS0yYTg3LTJkYTQtOWNjNC05NjFmZDExM2JlY2UiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNS43MTk1OTQ1NjZaIn19LCJiZGF5IjoiMjAyMi0wNS0xN1QwNjo0ODozOC4wMTc5ODgyMDZaIn0=; KRTBCOOKIE_153=1923-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&19420-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&22979-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&23462-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse; KRTBCOOKIE_57=22776-41928985301451193&KRTB&23339-41928985301451193; KRTBCOOKIE_27=16735-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&16736-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&23019-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&23114-uid:3cab6283-4546-4500-a7b6-40ef605fe745; KRTBCOOKIE_18=22947-1978557989514665832; KRTBCOOKIE_466=16530-4fc36250-d852-459c-8772-7356de17ab97; KRTBCOOKIE_391=22924-8044608333778839078&KRTB&23263-8044608333778839078&KRTB&23481-8044608333778839078; KRTBCOOKIE_1310=23431-b81c3g7dr67i&KRTB&23446-b81c3g7dr67i&KRTB&23465-b81c3g7dr67i; KRTBCOOKIE_1290=23368-vkf3yv9lbbl; KRTBCOOKIE_22=14911-4554572065121110164&KRTB&23150-4554572065121110164; KRTBCOOKIE_860=16335-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23334-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23417-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23426-YGAqDU1zUTdjyAFxCoe3kctlNPo; KRTBCOOKIE_904=16787-KwJwE7NkCZClNJRysN2iYg; KRTBCOOKIE_1159=23138-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23328-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23427-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23445-5545f53f3d6e4ec199d8ed627ff026f3; KRTBCOOKIE_32=11175-AQEI_1QecY2ESAIjEW6KAQEBAQE&KRTB&22713-AQEI_1QecY2ESAIjEW6KAQEBAQE&KRTB&22715-AQEI_1QecY2ESAIjEW6KAQEBAQE; SyncRTB3=1685577600%3A35%7C1685491200%3A107_21_71_56_204_247_165_231_233_179_22_209_54_254_238_96_99_220_7_214_13_3_8_234_176_46_5%7C1684886400%3A2_223_15%7C1689465600%3A69%7C1685145600%3A63; KRTBCOOKIE_107=1471-uid:EK38R0PM1NQR0H5&KRTB&23421-uid:EK38R0PM1NQR0H5; KRTBCOOKIE_594=17105-RX-447a6332-530e-456a-97f4-3f0fd1ed48c9-004&KRTB&17107-RX-447a6332-530e-456a-97f4-3f0fd1ed48c9-004; SPugT=1684310122; chkChromeAb67Sec=133; KRTBCOOKIE_699=22727-AAFy2k7FBosAAEasbJoXnw; PugT=1684310473; origin=go-test`)
						return r
					}(),
					Body: []byte(`{"ext":{"wrapper":{"profileid":5890,"versionid":1}}}`),
				},
				setup: func(mme *mock_metrics.MockMetricsEngine) {},
			},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{
				ModuleContext: hookstage.ModuleContext{
					"rctx": models.RequestCtx{
						Endpoint: models.EndpointHybrid,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "valid /openrtb2/auction?source=owsdk request directly on port 8000",
			fields: fields{
				cfg: config.Config{
					Tracker: config.Tracker{
						Endpoint:                  "t.pubmatic.com",
						VideoErrorTrackerEndpoint: "t.pubmatic.com/error",
					},
				},
				cache: nil,
			},
			args: args{
				in0:   context.Background(),
				miCtx: hookstage.ModuleInvocationContext{},
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						r, err := http.NewRequest("POST", "http://localhost/openrtb2/auction?source=owsdk&debug=1", nil)
						if err != nil {
							panic(err)
						}
						r.Header.Add("User-Agent", "go-test")
						r.Header.Add("SOURCE_IP", "127.0.0.1")
						r.Header.Add("Cookie", `KADUSERCOOKIE=7D75D25F-FAC9-443D-B2D1-B17FEE11E027; DPSync3=1684886400%3A248%7C1685491200%3A245_226_201; KRTBCOOKIE_80=16514-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&22987-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&23025-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&23386-CAESEMih0bN7ISRdZT8xX8LXzEw; KRTBCOOKIE_377=6810-59dc50c9-d658-44ce-b442-5a1f344d97c0&KRTB&22918-59dc50c9-d658-44ce-b442-5a1f344d97c0&KRTB&23031-59dc50c9-d658-44ce-b442-5a1f344d97c0; uids=eyJ0ZW1wVUlEcyI6eyIzM2Fjcm9zcyI6eyJ1aWQiOiIxMTkxNzkxMDk5Nzc2NjEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTo0My4zODg4Nzk5NVoifSwiYWRmIjp7InVpZCI6IjgwNDQ2MDgzMzM3Nzg4MzkwNzgiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMS4wMzMwNTQ3MjdaIn0sImFka2VybmVsIjp7InVpZCI6IkE5MTYzNTAwNzE0OTkyOTMyOTkwIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuMzczMzg1NjYyWiJ9LCJhZGtlcm5lbEFkbiI6eyJ1aWQiOiJBOTE2MzUwMDcxNDk5MjkzMjk5MCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEzLjQzNDkyNTg5NloifSwiYWRtaXhlciI6eyJ1aWQiOiIzNjZhMTdiMTJmMjI0ZDMwOGYzZTNiOGRhOGMzYzhhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjU5MjkxNDgwMVoifSwiYWRueHMiOnsidWlkIjoiNDE5Mjg5ODUzMDE0NTExOTMiLCJleHBpcmVzIjoiMjAyMy0wMS0xOFQwOTo1MzowOC44MjU0NDI2NzZaIn0sImFqYSI6eyJ1aWQiOiJzMnN1aWQ2RGVmMFl0bjJveGQ1aG9zS1AxVmV3IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTMuMjM5MTc2MDU0WiJ9LCJlcGxhbm5pbmciOnsidWlkIjoiQUoxRjBTOE5qdTdTQ0xWOSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjkyOTk2MDQ3M1oifSwiZ2Ftb3NoaSI6eyJ1aWQiOiJndXNyXzM1NmFmOWIxZDhjNjQyYjQ4MmNiYWQyYjdhMjg4MTYxIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuNTI0MTU3MjI1WiJ9LCJncmlkIjp7InVpZCI6IjRmYzM2MjUwLWQ4NTItNDU5Yy04NzcyLTczNTZkZTE3YWI5NyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE0LjY5NjMxNjIyN1oifSwiZ3JvdXBtIjp7InVpZCI6IjdENzVEMjVGLUZBQzktNDQzRC1CMkQxLUIxN0ZFRTExRTAyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjM5LjIyNjIxMjUzMloifSwiaXgiOnsidWlkIjoiWW9ORlNENlc5QkphOEh6eEdtcXlCUUFBXHUwMDI2Mjk3IiwiZXhwaXJlcyI6IjIwMjMtMDUtMzFUMDc6NTM6MzguNTU1ODI3MzU0WiJ9LCJqaXhpZSI6eyJ1aWQiOiI3MzY3MTI1MC1lODgyLTExZWMtYjUzOC0xM2FjYjdhZjBkZTQiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi4xOTEwOTk3MzJaIn0sImxvZ2ljYWQiOnsidWlkIjoiQVZ4OVROQS11c25pa3M4QURzTHpWa3JvaDg4QUFBR0JUREh0UUEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS40NTUxNDk2MTZaIn0sIm1lZGlhbmV0Ijp7InVpZCI6IjI5Nzg0MjM0OTI4OTU0MTAwMDBWMTAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMy42NzIyMTUxMjhaIn0sIm1naWQiOnsidWlkIjoibTU5Z1hyN0xlX1htIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTcuMDk3MDAxNDcxWiJ9LCJuYW5vaW50ZXJhY3RpdmUiOnsidWlkIjoiNmFlYzhjMTAzNzlkY2I3ODQxMmJjODBiNmRkOWM5NzMxNzNhYjdkNzEyZTQzMWE1YTVlYTcwMzRlNTZhNThhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE2LjcxNDgwNzUwNVoifSwib25ldGFnIjp7InVpZCI6IjdPelZoVzFOeC1LOGFVak1HMG52NXVNYm5YNEFHUXZQbnVHcHFrZ3k0ckEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS4xNDE3NDEyNjJaIn0sIm9wZW54Ijp7InVpZCI6IjVkZWNlNjIyLTBhMjMtMGRhYi0zYTI0LTVhNzcwMTBlNDU4MiIsImV4cGlyZXMiOiIyMDIzLTA1LTMxVDA3OjUyOjQ3LjE0MDQxNzM2M1oifSwicHVibWF0aWMiOnsidWlkIjoiN0Q3NUQyNUYtRkFDOS00NDNELUIyRDEtQjE3RkVFMTFFMDI3IiwiZXhwaXJlcyI6IjIwMjItMTAtMzFUMDk6MTQ6MjUuNzM3MjU2ODk5WiJ9LCJyaWNoYXVkaWVuY2UiOnsidWlkIjoiY2I2YzYzMjAtMzNlMi00Nzc0LWIxNjAtMXp6MTY1NDg0MDc0OSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjUyNTA3NDE4WiJ9LCJzbWFydHlhZHMiOnsidWlkIjoiMTJhZjE1ZTQ0ZjAwZDA3NjMwZTc0YzQ5MTU0Y2JmYmE0Zjg0N2U4ZDRhMTU0YzhjM2Q1MWY1OGNmNzJhNDYyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjgyNTAzMTg4NFoifSwic21pbGV3YW50ZWQiOnsidWlkIjoiZGQ5YzNmZTE4N2VmOWIwOWNhYTViNzExNDA0YzI4MzAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNC4yNTU2MDkzNjNaIn0sInN5bmFjb3JtZWRpYSI6eyJ1aWQiOiJHRFBSIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuOTc5NTgzNDM4WiJ9LCJ0cmlwbGVsaWZ0Ijp7InVpZCI6IjcwMjE5NzUwNTQ4MDg4NjUxOTQ2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA4Ljk4OTY3MzU3NFoifSwidmFsdWVpbXByZXNzaW9uIjp7InVpZCI6IjlkMDgxNTVmLWQ5ZmUtNGI1OC04OThlLWUyYzU2MjgyYWIzZSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjA2NzgzOTE2NFoifSwidmlzeCI6eyJ1aWQiOiIyN2UwYWMzYy1iNDZlLTQxYjMtOTkyYy1mOGQyNzE0OTQ5NWUiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi45ODk1MjM1NzNaIn0sInlpZWxkbGFiIjp7InVpZCI6IjY5NzE0ZDlmLWZiMDAtNGE1Zi04MTljLTRiZTE5MTM2YTMyNSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjExLjMwMzAyNjYxNVoifSwieWllbGRtbyI6eyJ1aWQiOiJnOTZjMmY3MTlmMTU1MWIzMWY2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjExMDUyODYwOVoifSwieWllbGRvbmUiOnsidWlkIjoiMmE0MmZiZDMtMmM3MC00ZWI5LWIxYmQtMDQ2OTY2NTBkOTQ4IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuMzE4MzMzOTM5WiJ9LCJ6ZXJvY2xpY2tmcmF1ZCI6eyJ1aWQiOiJiOTk5NThmZS0yYTg3LTJkYTQtOWNjNC05NjFmZDExM2JlY2UiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNS43MTk1OTQ1NjZaIn19LCJiZGF5IjoiMjAyMi0wNS0xN1QwNjo0ODozOC4wMTc5ODgyMDZaIn0=; KRTBCOOKIE_153=1923-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&19420-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&22979-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&23462-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse; KRTBCOOKIE_57=22776-41928985301451193&KRTB&23339-41928985301451193; KRTBCOOKIE_27=16735-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&16736-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&23019-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&23114-uid:3cab6283-4546-4500-a7b6-40ef605fe745; KRTBCOOKIE_18=22947-1978557989514665832; KRTBCOOKIE_466=16530-4fc36250-d852-459c-8772-7356de17ab97; KRTBCOOKIE_391=22924-8044608333778839078&KRTB&23263-8044608333778839078&KRTB&23481-8044608333778839078; KRTBCOOKIE_1310=23431-b81c3g7dr67i&KRTB&23446-b81c3g7dr67i&KRTB&23465-b81c3g7dr67i; KRTBCOOKIE_1290=23368-vkf3yv9lbbl; KRTBCOOKIE_22=14911-4554572065121110164&KRTB&23150-4554572065121110164; KRTBCOOKIE_860=16335-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23334-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23417-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23426-YGAqDU1zUTdjyAFxCoe3kctlNPo; KRTBCOOKIE_904=16787-KwJwE7NkCZClNJRysN2iYg; KRTBCOOKIE_1159=23138-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23328-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23427-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23445-5545f53f3d6e4ec199d8ed627ff026f3; KRTBCOOKIE_32=11175-AQEI_1QecY2ESAIjEW6KAQEBAQE&KRTB&22713-AQEI_1QecY2ESAIjEW6KAQEBAQE&KRTB&22715-AQEI_1QecY2ESAIjEW6KAQEBAQE; SyncRTB3=1685577600%3A35%7C1685491200%3A107_21_71_56_204_247_165_231_233_179_22_209_54_254_238_96_99_220_7_214_13_3_8_234_176_46_5%7C1684886400%3A2_223_15%7C1689465600%3A69%7C1685145600%3A63; KRTBCOOKIE_107=1471-uid:EK38R0PM1NQR0H5&KRTB&23421-uid:EK38R0PM1NQR0H5; KRTBCOOKIE_594=17105-RX-447a6332-530e-456a-97f4-3f0fd1ed48c9-004&KRTB&17107-RX-447a6332-530e-456a-97f4-3f0fd1ed48c9-004; SPugT=1684310122; chkChromeAb67Sec=133; KRTBCOOKIE_699=22727-AAFy2k7FBosAAEasbJoXnw; PugT=1684310473; origin=go-test`)
						return r
					}(),
					Body: []byte(`{"ext":{"wrapper":{"profileid":5890,"versionid":1}},"app":{"publisher":{"id":"5890"}}}`),
				},
				setup: func(mme *mock_metrics.MockMetricsEngine) {},
			},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{
				ModuleContext: hookstage.ModuleContext{
					"rctx": models.RequestCtx{
						PubIDStr:                  "5890",
						PubID:                     5890,
						ProfileID:                 5890,
						DisplayID:                 1,
						DisplayVersionID:          1,
						SSAuction:                 -1,
						Debug:                     true,
						DeviceCtx:                 models.DeviceCtx{IP: "127.0.0.1", UA: "go-test"},
						IsCTVRequest:              false,
						TrackerEndpoint:           "t.pubmatic.com",
						VideoErrorTrackerEndpoint: "t.pubmatic.com/error",
						UidCookie: &http.Cookie{
							Name:  "uids",
							Value: `eyJ0ZW1wVUlEcyI6eyIzM2Fjcm9zcyI6eyJ1aWQiOiIxMTkxNzkxMDk5Nzc2NjEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTo0My4zODg4Nzk5NVoifSwiYWRmIjp7InVpZCI6IjgwNDQ2MDgzMzM3Nzg4MzkwNzgiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMS4wMzMwNTQ3MjdaIn0sImFka2VybmVsIjp7InVpZCI6IkE5MTYzNTAwNzE0OTkyOTMyOTkwIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuMzczMzg1NjYyWiJ9LCJhZGtlcm5lbEFkbiI6eyJ1aWQiOiJBOTE2MzUwMDcxNDk5MjkzMjk5MCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEzLjQzNDkyNTg5NloifSwiYWRtaXhlciI6eyJ1aWQiOiIzNjZhMTdiMTJmMjI0ZDMwOGYzZTNiOGRhOGMzYzhhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjU5MjkxNDgwMVoifSwiYWRueHMiOnsidWlkIjoiNDE5Mjg5ODUzMDE0NTExOTMiLCJleHBpcmVzIjoiMjAyMy0wMS0xOFQwOTo1MzowOC44MjU0NDI2NzZaIn0sImFqYSI6eyJ1aWQiOiJzMnN1aWQ2RGVmMFl0bjJveGQ1aG9zS1AxVmV3IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTMuMjM5MTc2MDU0WiJ9LCJlcGxhbm5pbmciOnsidWlkIjoiQUoxRjBTOE5qdTdTQ0xWOSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjkyOTk2MDQ3M1oifSwiZ2Ftb3NoaSI6eyJ1aWQiOiJndXNyXzM1NmFmOWIxZDhjNjQyYjQ4MmNiYWQyYjdhMjg4MTYxIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuNTI0MTU3MjI1WiJ9LCJncmlkIjp7InVpZCI6IjRmYzM2MjUwLWQ4NTItNDU5Yy04NzcyLTczNTZkZTE3YWI5NyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE0LjY5NjMxNjIyN1oifSwiZ3JvdXBtIjp7InVpZCI6IjdENzVEMjVGLUZBQzktNDQzRC1CMkQxLUIxN0ZFRTExRTAyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjM5LjIyNjIxMjUzMloifSwiaXgiOnsidWlkIjoiWW9ORlNENlc5QkphOEh6eEdtcXlCUUFBXHUwMDI2Mjk3IiwiZXhwaXJlcyI6IjIwMjMtMDUtMzFUMDc6NTM6MzguNTU1ODI3MzU0WiJ9LCJqaXhpZSI6eyJ1aWQiOiI3MzY3MTI1MC1lODgyLTExZWMtYjUzOC0xM2FjYjdhZjBkZTQiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi4xOTEwOTk3MzJaIn0sImxvZ2ljYWQiOnsidWlkIjoiQVZ4OVROQS11c25pa3M4QURzTHpWa3JvaDg4QUFBR0JUREh0UUEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS40NTUxNDk2MTZaIn0sIm1lZGlhbmV0Ijp7InVpZCI6IjI5Nzg0MjM0OTI4OTU0MTAwMDBWMTAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMy42NzIyMTUxMjhaIn0sIm1naWQiOnsidWlkIjoibTU5Z1hyN0xlX1htIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTcuMDk3MDAxNDcxWiJ9LCJuYW5vaW50ZXJhY3RpdmUiOnsidWlkIjoiNmFlYzhjMTAzNzlkY2I3ODQxMmJjODBiNmRkOWM5NzMxNzNhYjdkNzEyZTQzMWE1YTVlYTcwMzRlNTZhNThhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE2LjcxNDgwNzUwNVoifSwib25ldGFnIjp7InVpZCI6IjdPelZoVzFOeC1LOGFVak1HMG52NXVNYm5YNEFHUXZQbnVHcHFrZ3k0ckEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS4xNDE3NDEyNjJaIn0sIm9wZW54Ijp7InVpZCI6IjVkZWNlNjIyLTBhMjMtMGRhYi0zYTI0LTVhNzcwMTBlNDU4MiIsImV4cGlyZXMiOiIyMDIzLTA1LTMxVDA3OjUyOjQ3LjE0MDQxNzM2M1oifSwicHVibWF0aWMiOnsidWlkIjoiN0Q3NUQyNUYtRkFDOS00NDNELUIyRDEtQjE3RkVFMTFFMDI3IiwiZXhwaXJlcyI6IjIwMjItMTAtMzFUMDk6MTQ6MjUuNzM3MjU2ODk5WiJ9LCJyaWNoYXVkaWVuY2UiOnsidWlkIjoiY2I2YzYzMjAtMzNlMi00Nzc0LWIxNjAtMXp6MTY1NDg0MDc0OSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjUyNTA3NDE4WiJ9LCJzbWFydHlhZHMiOnsidWlkIjoiMTJhZjE1ZTQ0ZjAwZDA3NjMwZTc0YzQ5MTU0Y2JmYmE0Zjg0N2U4ZDRhMTU0YzhjM2Q1MWY1OGNmNzJhNDYyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjgyNTAzMTg4NFoifSwic21pbGV3YW50ZWQiOnsidWlkIjoiZGQ5YzNmZTE4N2VmOWIwOWNhYTViNzExNDA0YzI4MzAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNC4yNTU2MDkzNjNaIn0sInN5bmFjb3JtZWRpYSI6eyJ1aWQiOiJHRFBSIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuOTc5NTgzNDM4WiJ9LCJ0cmlwbGVsaWZ0Ijp7InVpZCI6IjcwMjE5NzUwNTQ4MDg4NjUxOTQ2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA4Ljk4OTY3MzU3NFoifSwidmFsdWVpbXByZXNzaW9uIjp7InVpZCI6IjlkMDgxNTVmLWQ5ZmUtNGI1OC04OThlLWUyYzU2MjgyYWIzZSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjA2NzgzOTE2NFoifSwidmlzeCI6eyJ1aWQiOiIyN2UwYWMzYy1iNDZlLTQxYjMtOTkyYy1mOGQyNzE0OTQ5NWUiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi45ODk1MjM1NzNaIn0sInlpZWxkbGFiIjp7InVpZCI6IjY5NzE0ZDlmLWZiMDAtNGE1Zi04MTljLTRiZTE5MTM2YTMyNSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjExLjMwMzAyNjYxNVoifSwieWllbGRtbyI6eyJ1aWQiOiJnOTZjMmY3MTlmMTU1MWIzMWY2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjExMDUyODYwOVoifSwieWllbGRvbmUiOnsidWlkIjoiMmE0MmZiZDMtMmM3MC00ZWI5LWIxYmQtMDQ2OTY2NTBkOTQ4IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuMzE4MzMzOTM5WiJ9LCJ6ZXJvY2xpY2tmcmF1ZCI6eyJ1aWQiOiJiOTk5NThmZS0yYTg3LTJkYTQtOWNjNC05NjFmZDExM2JlY2UiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNS43MTk1OTQ1NjZaIn19LCJiZGF5IjoiMjAyMi0wNS0xN1QwNjo0ODozOC4wMTc5ODgyMDZaIn0=`,
						},
						KADUSERCookie: &http.Cookie{
							Name:  "KADUSERCOOKIE",
							Value: `7D75D25F-FAC9-443D-B2D1-B17FEE11E027`,
						},
						OriginCookie:                    "go-test",
						Aliases:                         make(map[string]string),
						ImpBidCtx:                       make(map[string]models.ImpCtx),
						PrebidBidderCode:                make(map[string]string),
						BidderResponseTimeMillis:        make(map[string]int),
						ProfileIDStr:                    "5890",
						Endpoint:                        models.EndpointV25,
						MetricsEngine:                   mockEngine,
						SeatNonBids:                     make(map[string][]openrtb_ext.NonBid),
						Method:                          "POST",
						WakandaDebug:                    &wakanda.Debug{},
						ImpCountingMethodEnabledBidders: make(map[string]struct{}),
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "valid request for applovin max /openrtb2/auction?source=owsdk&agent=max request",
			fields: fields{
				cfg: config.Config{
					Tracker: config.Tracker{
						Endpoint:                  "t.pubmatic.com",
						VideoErrorTrackerEndpoint: "t.pubmatic.com/error",
					},
				},
				cache: nil,
			},
			args: args{
				in0:   context.Background(),
				miCtx: hookstage.ModuleInvocationContext{},
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						r, err := http.NewRequest("POST", "http://localhost/openrtb2/auction?source=owsdk&agent=max&debug=1", nil)
						if err != nil {
							panic(err)
						}
						r.Header.Add("User-Agent", "go-test")
						r.Header.Add("SOURCE_IP", "127.0.0.1")
						r.Header.Add("Cookie", `KADUSERCOOKIE=7D75D25F-FAC9-443D-B2D1-B17FEE11E027; DPSync3=1684886400%3A248%7C1685491200%3A245_226_201; KRTBCOOKIE_80=16514-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&22987-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&23025-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&23386-CAESEMih0bN7ISRdZT8xX8LXzEw; KRTBCOOKIE_377=6810-59dc50c9-d658-44ce-b442-5a1f344d97c0&KRTB&22918-59dc50c9-d658-44ce-b442-5a1f344d97c0&KRTB&23031-59dc50c9-d658-44ce-b442-5a1f344d97c0; uids=eyJ0ZW1wVUlEcyI6eyIzM2Fjcm9zcyI6eyJ1aWQiOiIxMTkxNzkxMDk5Nzc2NjEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTo0My4zODg4Nzk5NVoifSwiYWRmIjp7InVpZCI6IjgwNDQ2MDgzMzM3Nzg4MzkwNzgiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMS4wMzMwNTQ3MjdaIn0sImFka2VybmVsIjp7InVpZCI6IkE5MTYzNTAwNzE0OTkyOTMyOTkwIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuMzczMzg1NjYyWiJ9LCJhZGtlcm5lbEFkbiI6eyJ1aWQiOiJBOTE2MzUwMDcxNDk5MjkzMjk5MCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEzLjQzNDkyNTg5NloifSwiYWRtaXhlciI6eyJ1aWQiOiIzNjZhMTdiMTJmMjI0ZDMwOGYzZTNiOGRhOGMzYzhhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjU5MjkxNDgwMVoifSwiYWRueHMiOnsidWlkIjoiNDE5Mjg5ODUzMDE0NTExOTMiLCJleHBpcmVzIjoiMjAyMy0wMS0xOFQwOTo1MzowOC44MjU0NDI2NzZaIn0sImFqYSI6eyJ1aWQiOiJzMnN1aWQ2RGVmMFl0bjJveGQ1aG9zS1AxVmV3IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTMuMjM5MTc2MDU0WiJ9LCJlcGxhbm5pbmciOnsidWlkIjoiQUoxRjBTOE5qdTdTQ0xWOSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjkyOTk2MDQ3M1oifSwiZ2Ftb3NoaSI6eyJ1aWQiOiJndXNyXzM1NmFmOWIxZDhjNjQyYjQ4MmNiYWQyYjdhMjg4MTYxIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuNTI0MTU3MjI1WiJ9LCJncmlkIjp7InVpZCI6IjRmYzM2MjUwLWQ4NTItNDU5Yy04NzcyLTczNTZkZTE3YWI5NyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE0LjY5NjMxNjIyN1oifSwiZ3JvdXBtIjp7InVpZCI6IjdENzVEMjVGLUZBQzktNDQzRC1CMkQxLUIxN0ZFRTExRTAyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjM5LjIyNjIxMjUzMloifSwiaXgiOnsidWlkIjoiWW9ORlNENlc5QkphOEh6eEdtcXlCUUFBXHUwMDI2Mjk3IiwiZXhwaXJlcyI6IjIwMjMtMDUtMzFUMDc6NTM6MzguNTU1ODI3MzU0WiJ9LCJqaXhpZSI6eyJ1aWQiOiI3MzY3MTI1MC1lODgyLTExZWMtYjUzOC0xM2FjYjdhZjBkZTQiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi4xOTEwOTk3MzJaIn0sImxvZ2ljYWQiOnsidWlkIjoiQVZ4OVROQS11c25pa3M4QURzTHpWa3JvaDg4QUFBR0JUREh0UUEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS40NTUxNDk2MTZaIn0sIm1lZGlhbmV0Ijp7InVpZCI6IjI5Nzg0MjM0OTI4OTU0MTAwMDBWMTAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMy42NzIyMTUxMjhaIn0sIm1naWQiOnsidWlkIjoibTU5Z1hyN0xlX1htIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTcuMDk3MDAxNDcxWiJ9LCJuYW5vaW50ZXJhY3RpdmUiOnsidWlkIjoiNmFlYzhjMTAzNzlkY2I3ODQxMmJjODBiNmRkOWM5NzMxNzNhYjdkNzEyZTQzMWE1YTVlYTcwMzRlNTZhNThhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE2LjcxNDgwNzUwNVoifSwib25ldGFnIjp7InVpZCI6IjdPelZoVzFOeC1LOGFVak1HMG52NXVNYm5YNEFHUXZQbnVHcHFrZ3k0ckEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS4xNDE3NDEyNjJaIn0sIm9wZW54Ijp7InVpZCI6IjVkZWNlNjIyLTBhMjMtMGRhYi0zYTI0LTVhNzcwMTBlNDU4MiIsImV4cGlyZXMiOiIyMDIzLTA1LTMxVDA3OjUyOjQ3LjE0MDQxNzM2M1oifSwicHVibWF0aWMiOnsidWlkIjoiN0Q3NUQyNUYtRkFDOS00NDNELUIyRDEtQjE3RkVFMTFFMDI3IiwiZXhwaXJlcyI6IjIwMjItMTAtMzFUMDk6MTQ6MjUuNzM3MjU2ODk5WiJ9LCJyaWNoYXVkaWVuY2UiOnsidWlkIjoiY2I2YzYzMjAtMzNlMi00Nzc0LWIxNjAtMXp6MTY1NDg0MDc0OSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjUyNTA3NDE4WiJ9LCJzbWFydHlhZHMiOnsidWlkIjoiMTJhZjE1ZTQ0ZjAwZDA3NjMwZTc0YzQ5MTU0Y2JmYmE0Zjg0N2U4ZDRhMTU0YzhjM2Q1MWY1OGNmNzJhNDYyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjgyNTAzMTg4NFoifSwic21pbGV3YW50ZWQiOnsidWlkIjoiZGQ5YzNmZTE4N2VmOWIwOWNhYTViNzExNDA0YzI4MzAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNC4yNTU2MDkzNjNaIn0sInN5bmFjb3JtZWRpYSI6eyJ1aWQiOiJHRFBSIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuOTc5NTgzNDM4WiJ9LCJ0cmlwbGVsaWZ0Ijp7InVpZCI6IjcwMjE5NzUwNTQ4MDg4NjUxOTQ2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA4Ljk4OTY3MzU3NFoifSwidmFsdWVpbXByZXNzaW9uIjp7InVpZCI6IjlkMDgxNTVmLWQ5ZmUtNGI1OC04OThlLWUyYzU2MjgyYWIzZSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjA2NzgzOTE2NFoifSwidmlzeCI6eyJ1aWQiOiIyN2UwYWMzYy1iNDZlLTQxYjMtOTkyYy1mOGQyNzE0OTQ5NWUiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi45ODk1MjM1NzNaIn0sInlpZWxkbGFiIjp7InVpZCI6IjY5NzE0ZDlmLWZiMDAtNGE1Zi04MTljLTRiZTE5MTM2YTMyNSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjExLjMwMzAyNjYxNVoifSwieWllbGRtbyI6eyJ1aWQiOiJnOTZjMmY3MTlmMTU1MWIzMWY2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjExMDUyODYwOVoifSwieWllbGRvbmUiOnsidWlkIjoiMmE0MmZiZDMtMmM3MC00ZWI5LWIxYmQtMDQ2OTY2NTBkOTQ4IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuMzE4MzMzOTM5WiJ9LCJ6ZXJvY2xpY2tmcmF1ZCI6eyJ1aWQiOiJiOTk5NThmZS0yYTg3LTJkYTQtOWNjNC05NjFmZDExM2JlY2UiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNS43MTk1OTQ1NjZaIn19LCJiZGF5IjoiMjAyMi0wNS0xN1QwNjo0ODozOC4wMTc5ODgyMDZaIn0=; KRTBCOOKIE_153=1923-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&19420-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&22979-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&23462-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse; KRTBCOOKIE_57=22776-41928985301451193&KRTB&23339-41928985301451193; KRTBCOOKIE_27=16735-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&16736-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&23019-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&23114-uid:3cab6283-4546-4500-a7b6-40ef605fe745; KRTBCOOKIE_18=22947-1978557989514665832; KRTBCOOKIE_466=16530-4fc36250-d852-459c-8772-7356de17ab97; KRTBCOOKIE_391=22924-8044608333778839078&KRTB&23263-8044608333778839078&KRTB&23481-8044608333778839078; KRTBCOOKIE_1310=23431-b81c3g7dr67i&KRTB&23446-b81c3g7dr67i&KRTB&23465-b81c3g7dr67i; KRTBCOOKIE_1290=23368-vkf3yv9lbbl; KRTBCOOKIE_22=14911-4554572065121110164&KRTB&23150-4554572065121110164; KRTBCOOKIE_860=16335-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23334-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23417-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23426-YGAqDU1zUTdjyAFxCoe3kctlNPo; KRTBCOOKIE_904=16787-KwJwE7NkCZClNJRysN2iYg; KRTBCOOKIE_1159=23138-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23328-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23427-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23445-5545f53f3d6e4ec199d8ed627ff026f3; KRTBCOOKIE_32=11175-AQEI_1QecY2ESAIjEW6KAQEBAQE&KRTB&22713-AQEI_1QecY2ESAIjEW6KAQEBAQE&KRTB&22715-AQEI_1QecY2ESAIjEW6KAQEBAQE; SyncRTB3=1685577600%3A35%7C1685491200%3A107_21_71_56_204_247_165_231_233_179_22_209_54_254_238_96_99_220_7_214_13_3_8_234_176_46_5%7C1684886400%3A2_223_15%7C1689465600%3A69%7C1685145600%3A63; KRTBCOOKIE_107=1471-uid:EK38R0PM1NQR0H5&KRTB&23421-uid:EK38R0PM1NQR0H5; KRTBCOOKIE_594=17105-RX-447a6332-530e-456a-97f4-3f0fd1ed48c9-004&KRTB&17107-RX-447a6332-530e-456a-97f4-3f0fd1ed48c9-004; SPugT=1684310122; chkChromeAb67Sec=133; KRTBCOOKIE_699=22727-AAFy2k7FBosAAEasbJoXnw; PugT=1684310473; origin=go-test`)
						return r
					}(),
					Body: []byte(`{"id":"test-case-1","at":1,"bcat":["IAB26-4","IAB26-2","IAB25-6","IAB25-5","IAB25-4","IAB25-3","IAB25-1","IAB25-7","IAB8-18","IAB26-3","IAB26-1","IAB8-5","IAB25-2","IAB11-4"],"tmax":1000,"app":{"publisher":{"name":"New Story Inc.","id":"5890","ext":{"installed_sdk":{"id":"MOLOCO_BIDDING","sdk_version":{"major":1,"minor":0,"micro":0},"adapter_version":{"major":1,"minor":0,"micro":0}}}},"paid":0,"name":"DrawHappyAngel","ver":"0.5.4","bundle":"com.newstory.DrawHappyAngel","cat":["IAB9-30"],"id":"1234567","ext":{"orientation":1}},"device":{"ifa":"497a10d6-c4dd-4e04-a986-c32b7180d462","ip":"38.158.207.171","carrier":"MYTEL","language":"en_US","hwv":"ruby","ppi":440,"pxratio":2.75,"devicetype":4,"connectiontype":2,"js":1,"h":2400,"w":1080,"geo":{"type":2,"ipservice":3,"lat":40.7429,"lon":-73.9392,"long":-73.9392,"city":"Queens","country":"USA","region":"ny","dma":"501","metro":"501","zip":"11101","ext":{"org":"Myanmar Broadband Telecom Co.","isp":"Myanmar Broadband Telecom Co."}},"ext":{},"osv":"13.0.0","ua":"Mozilla/5.0 (Linux; Android 13; 22101316C Build/TP1A.220624.014; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.230 Mobile Safari/537.36","make":"xiaomi","model":"22101316c","os":"android"},"imp":[{"id":"1","displaymanager":"applovin_mediation","displaymanagerver":"11.8.2","instl":0,"secure":0,"tagid":"/43743431/DMDemo","bidfloor":0.01,"bidfloorcur":"USD","exp":14400,"banner":{"format":[{"w":728,"h":90},{"w":300,"h":250}],"w":700,"h":900},"video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"rwdd":0}],"user":{"data":[{"id":"1","name":"Publisher Passed","segment":[{"signal":"{\"id\":\"95d6643c-3da6-40a2-b9ca-12279393ffbf\",\"at\":1,\"tmax\":500,\"cur\":[\"USD\"],\"imp\":[{\"id\":\"imp176227948\",\"clickbrowser\":0,\"displaymanager\":\"PubMatic_OpenBid_SDK\",\"displaymanagerver\":\"1.4.0\",\"tagid\":\"/43743431/DMDemo\",\"secure\":0,\"banner\":{\"pos\":7,\"format\":[{\"w\":300,\"h\":250}],\"api\":[5,6,7]},\"instl\":1}],\"app\":{\"paid\":4,\"name\":\"OpenWrapperSample\",\"bundle\":\"com.pubmatic.openbid.app\",\"storeurl\":\"https://itunes.apple.com/us/app/pubmatic-sdk-app/id1175273098\",\"ver\":\"1.0\",\"publisher\":{\"id\":\"5890\"}},\"device\":{\"geo\":{\"type\":1,\"lat\":37.421998333333335,\"lon\":-122.08400000000002},\"pxratio\":2.625,\"mccmnc\":\"310-260\",\"lmt\":0,\"ifa\":\"07c387f2-e030-428f-8336-42f682150759\",\"connectiontype\":5,\"carrier\":\"Android\",\"js\":1,\"ua\":\"signal_UA\",\"make\":\"Google\",\"model\":\"AndroidSDKbuiltforx86\",\"os\":\"Android\",\"osv\":\"9\",\"h\":1794,\"w\":1080,\"language\":\"en-US\",\"devicetype\":4,\"ext\":{\"atts\":3}},\"source\":{\"ext\":{\"omidpn\":\"PubMatic\",\"omidpv\":\"1.2.11-Pubmatic\"}},\"user\":{\"data\":[{\"id\":\"1234\"}]},\"ext\":{\"wrapper\":{\"ssauction\":1,\"sumry_disable\":0,\"profileid\":58135,\"versionid\":1,\"clientconfig\":1}}}"}]}],"ext":{"gdpr":0}},"regs":{"coppa":0,"ext":{"gdpr":0}},"source":{"ext":{"schain":{"ver":"1.0","complete":1,"nodes":[{"asi":"applovin.com","sid":"53bf468f18c5a0e2b7d4e3f748c677c1","rid":"494dbe15a3ce08c54f4e456363f35a022247f997","hp":1}]}}},"ext":{"prebid":{"bidderparams":{"pubmatic":{"sendburl":true,"wrapper":{"profileid":12929,"versionid":1}}}}}}`),
				},
				setup: func(mme *mock_metrics.MockMetricsEngine) {
					mockFeature.EXPECT().GetApplovinSchainABTestPercentage().Return(10)
					mockEngine.EXPECT().RecordRequestWithSchainABTestEnabled()
					GetRandomNumberIn1To100 = func() int {
						return 5
					}
				},
			},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{
				ModuleContext: hookstage.ModuleContext{
					"rctx": models.RequestCtx{
						PubIDStr:                  "5890",
						PubID:                     5890,
						ProfileID:                 12929,
						DisplayID:                 1,
						DisplayVersionID:          1,
						SSAuction:                 -1,
						ClientConfigFlag:          1,
						Debug:                     true,
						DeviceCtx:                 models.DeviceCtx{IP: "127.0.0.1", UA: "go-test"},
						IsCTVRequest:              false,
						TrackerEndpoint:           "t.pubmatic.com",
						VideoErrorTrackerEndpoint: "t.pubmatic.com/error",
						UidCookie: &http.Cookie{
							Name:  "uids",
							Value: `eyJ0ZW1wVUlEcyI6eyIzM2Fjcm9zcyI6eyJ1aWQiOiIxMTkxNzkxMDk5Nzc2NjEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTo0My4zODg4Nzk5NVoifSwiYWRmIjp7InVpZCI6IjgwNDQ2MDgzMzM3Nzg4MzkwNzgiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMS4wMzMwNTQ3MjdaIn0sImFka2VybmVsIjp7InVpZCI6IkE5MTYzNTAwNzE0OTkyOTMyOTkwIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuMzczMzg1NjYyWiJ9LCJhZGtlcm5lbEFkbiI6eyJ1aWQiOiJBOTE2MzUwMDcxNDk5MjkzMjk5MCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEzLjQzNDkyNTg5NloifSwiYWRtaXhlciI6eyJ1aWQiOiIzNjZhMTdiMTJmMjI0ZDMwOGYzZTNiOGRhOGMzYzhhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjU5MjkxNDgwMVoifSwiYWRueHMiOnsidWlkIjoiNDE5Mjg5ODUzMDE0NTExOTMiLCJleHBpcmVzIjoiMjAyMy0wMS0xOFQwOTo1MzowOC44MjU0NDI2NzZaIn0sImFqYSI6eyJ1aWQiOiJzMnN1aWQ2RGVmMFl0bjJveGQ1aG9zS1AxVmV3IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTMuMjM5MTc2MDU0WiJ9LCJlcGxhbm5pbmciOnsidWlkIjoiQUoxRjBTOE5qdTdTQ0xWOSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjkyOTk2MDQ3M1oifSwiZ2Ftb3NoaSI6eyJ1aWQiOiJndXNyXzM1NmFmOWIxZDhjNjQyYjQ4MmNiYWQyYjdhMjg4MTYxIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuNTI0MTU3MjI1WiJ9LCJncmlkIjp7InVpZCI6IjRmYzM2MjUwLWQ4NTItNDU5Yy04NzcyLTczNTZkZTE3YWI5NyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE0LjY5NjMxNjIyN1oifSwiZ3JvdXBtIjp7InVpZCI6IjdENzVEMjVGLUZBQzktNDQzRC1CMkQxLUIxN0ZFRTExRTAyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjM5LjIyNjIxMjUzMloifSwiaXgiOnsidWlkIjoiWW9ORlNENlc5QkphOEh6eEdtcXlCUUFBXHUwMDI2Mjk3IiwiZXhwaXJlcyI6IjIwMjMtMDUtMzFUMDc6NTM6MzguNTU1ODI3MzU0WiJ9LCJqaXhpZSI6eyJ1aWQiOiI3MzY3MTI1MC1lODgyLTExZWMtYjUzOC0xM2FjYjdhZjBkZTQiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi4xOTEwOTk3MzJaIn0sImxvZ2ljYWQiOnsidWlkIjoiQVZ4OVROQS11c25pa3M4QURzTHpWa3JvaDg4QUFBR0JUREh0UUEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS40NTUxNDk2MTZaIn0sIm1lZGlhbmV0Ijp7InVpZCI6IjI5Nzg0MjM0OTI4OTU0MTAwMDBWMTAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMy42NzIyMTUxMjhaIn0sIm1naWQiOnsidWlkIjoibTU5Z1hyN0xlX1htIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTcuMDk3MDAxNDcxWiJ9LCJuYW5vaW50ZXJhY3RpdmUiOnsidWlkIjoiNmFlYzhjMTAzNzlkY2I3ODQxMmJjODBiNmRkOWM5NzMxNzNhYjdkNzEyZTQzMWE1YTVlYTcwMzRlNTZhNThhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE2LjcxNDgwNzUwNVoifSwib25ldGFnIjp7InVpZCI6IjdPelZoVzFOeC1LOGFVak1HMG52NXVNYm5YNEFHUXZQbnVHcHFrZ3k0ckEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS4xNDE3NDEyNjJaIn0sIm9wZW54Ijp7InVpZCI6IjVkZWNlNjIyLTBhMjMtMGRhYi0zYTI0LTVhNzcwMTBlNDU4MiIsImV4cGlyZXMiOiIyMDIzLTA1LTMxVDA3OjUyOjQ3LjE0MDQxNzM2M1oifSwicHVibWF0aWMiOnsidWlkIjoiN0Q3NUQyNUYtRkFDOS00NDNELUIyRDEtQjE3RkVFMTFFMDI3IiwiZXhwaXJlcyI6IjIwMjItMTAtMzFUMDk6MTQ6MjUuNzM3MjU2ODk5WiJ9LCJyaWNoYXVkaWVuY2UiOnsidWlkIjoiY2I2YzYzMjAtMzNlMi00Nzc0LWIxNjAtMXp6MTY1NDg0MDc0OSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjUyNTA3NDE4WiJ9LCJzbWFydHlhZHMiOnsidWlkIjoiMTJhZjE1ZTQ0ZjAwZDA3NjMwZTc0YzQ5MTU0Y2JmYmE0Zjg0N2U4ZDRhMTU0YzhjM2Q1MWY1OGNmNzJhNDYyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjgyNTAzMTg4NFoifSwic21pbGV3YW50ZWQiOnsidWlkIjoiZGQ5YzNmZTE4N2VmOWIwOWNhYTViNzExNDA0YzI4MzAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNC4yNTU2MDkzNjNaIn0sInN5bmFjb3JtZWRpYSI6eyJ1aWQiOiJHRFBSIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuOTc5NTgzNDM4WiJ9LCJ0cmlwbGVsaWZ0Ijp7InVpZCI6IjcwMjE5NzUwNTQ4MDg4NjUxOTQ2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA4Ljk4OTY3MzU3NFoifSwidmFsdWVpbXByZXNzaW9uIjp7InVpZCI6IjlkMDgxNTVmLWQ5ZmUtNGI1OC04OThlLWUyYzU2MjgyYWIzZSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjA2NzgzOTE2NFoifSwidmlzeCI6eyJ1aWQiOiIyN2UwYWMzYy1iNDZlLTQxYjMtOTkyYy1mOGQyNzE0OTQ5NWUiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi45ODk1MjM1NzNaIn0sInlpZWxkbGFiIjp7InVpZCI6IjY5NzE0ZDlmLWZiMDAtNGE1Zi04MTljLTRiZTE5MTM2YTMyNSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjExLjMwMzAyNjYxNVoifSwieWllbGRtbyI6eyJ1aWQiOiJnOTZjMmY3MTlmMTU1MWIzMWY2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjExMDUyODYwOVoifSwieWllbGRvbmUiOnsidWlkIjoiMmE0MmZiZDMtMmM3MC00ZWI5LWIxYmQtMDQ2OTY2NTBkOTQ4IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuMzE4MzMzOTM5WiJ9LCJ6ZXJvY2xpY2tmcmF1ZCI6eyJ1aWQiOiJiOTk5NThmZS0yYTg3LTJkYTQtOWNjNC05NjFmZDExM2JlY2UiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNS43MTk1OTQ1NjZaIn19LCJiZGF5IjoiMjAyMi0wNS0xN1QwNjo0ODozOC4wMTc5ODgyMDZaIn0=`,
						},
						KADUSERCookie: &http.Cookie{
							Name:  "KADUSERCOOKIE",
							Value: `7D75D25F-FAC9-443D-B2D1-B17FEE11E027`,
						},
						OriginCookie:                    "go-test",
						Aliases:                         make(map[string]string),
						ImpBidCtx:                       make(map[string]models.ImpCtx),
						PrebidBidderCode:                make(map[string]string),
						BidderResponseTimeMillis:        make(map[string]int),
						ProfileIDStr:                    "12929",
						Endpoint:                        models.EndpointAppLovinMax,
						MetricsEngine:                   mockEngine,
						SeatNonBids:                     make(map[string][]openrtb_ext.NonBid),
						Method:                          "POST",
						WakandaDebug:                    &wakanda.Debug{},
						SendBurl:                        true,
						ImpCountingMethodEnabledBidders: make(map[string]struct{}),
					},
				},
			},
			doMutate: true,
			wantErr:  nil,
			wantBody: []byte(`{"id":"test-case-1","imp":[{"id":"1","banner":{"format":[{"w":728,"h":90},{"w":300,"h":250}],"w":700,"h":900,"api":[5,6,7]},"video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"displaymanager":"PubMatic_OpenBid_SDK","displaymanagerver":"1.4.0","tagid":"/43743431/DMDemo","bidfloor":0.01,"bidfloorcur":"USD","clickbrowser":0,"secure":0,"exp":14400}],"app":{"id":"1234567","name":"DrawHappyAngel","bundle":"com.newstory.DrawHappyAngel","storeurl":"https://itunes.apple.com/us/app/pubmatic-sdk-app/id1175273098","cat":["IAB9-30"],"ver":"0.5.4","paid":4,"publisher":{"id":"5890","name":"New Story Inc.","ext":{"installed_sdk":{"id":"MOLOCO_BIDDING","sdk_version":{"major":1,"minor":0,"micro":0},"adapter_version":{"major":1,"minor":0,"micro":0}}}},"ext":{"orientation":1}},"device":{"geo":{"lat":40.7429,"lon":-73.9392,"type":2,"ipservice":3,"country":"USA","region":"ny","metro":"501","city":"Queens","zip":"11101","ext":{"org":"Myanmar Broadband Telecom Co.","isp":"Myanmar Broadband Telecom Co."}},"ua":"signal_UA","ip":"38.158.207.171","devicetype":4,"make":"xiaomi","model":"AndroidSDKbuiltforx86","os":"android","osv":"13.0.0","hwv":"ruby","h":2400,"w":1080,"ppi":440,"pxratio":2.75,"js":1,"language":"en_US","carrier":"MYTEL","mccmnc":"310-260","connectiontype":5,"ifa":"497a10d6-c4dd-4e04-a986-c32b7180d462","ext":{"atts":3}},"user":{"data":[{"id":"1234"}],"ext":{"gdpr":0}},"at":1,"tmax":1000,"bcat":["IAB26-4","IAB26-2","IAB25-6","IAB25-5","IAB25-4","IAB25-3","IAB25-1","IAB25-7","IAB8-18","IAB26-3","IAB26-1","IAB8-5","IAB25-2","IAB11-4"],"source":{"ext":{"schain":{"ver":"1.0","complete":1,"nodes":[{"asi":"applovin.com","sid":"53bf468f18c5a0e2b7d4e3f748c677c1","rid":"494dbe15a3ce08c54f4e456363f35a022247f997","hp":1}]},"omidpn":"PubMatic","omidpv":"1.2.11-Pubmatic"}},"regs":{"ext":{"gdpr":0}},"ext":{"prebid":{"bidderparams":{"pubmatic":{"sendburl":true,"wrapper":{"profileid":12929,"versionid":1,"clientconfig":1}}}}}}`),
		},
		{
			name: "valid request for applovin max /openrtb2/auction?source=owsdk&agent=max request with wakanda enable",
			fields: fields{
				cfg: config.Config{
					Tracker: config.Tracker{
						Endpoint:                  "t.pubmatic.com",
						VideoErrorTrackerEndpoint: "t.pubmatic.com/error",
					},
				},
				cache: nil,
			},
			args: args{
				in0:   context.Background(),
				miCtx: hookstage.ModuleInvocationContext{},
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						r, err := http.NewRequest("POST", "http://localhost/openrtb2/auction?source=owsdk&agent=max&debug=1", nil)
						if err != nil {
							panic(err)
						}
						r.Header.Add("User-Agent", "go-test")
						r.Header.Add("Source_ip", "127.0.0.1")
						r.Header.Add("Cookie", `KADUSERCOOKIE=7D75D25F-FAC9-443D-B2D1-B17FEE11E027; DPSync3=1684886400%3A248%7C1685491200%3A245_226_201; KRTBCOOKIE_80=16514-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&22987-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&23025-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&23386-CAESEMih0bN7ISRdZT8xX8LXzEw; KRTBCOOKIE_377=6810-59dc50c9-d658-44ce-b442-5a1f344d97c0&KRTB&22918-59dc50c9-d658-44ce-b442-5a1f344d97c0&KRTB&23031-59dc50c9-d658-44ce-b442-5a1f344d97c0; uids=eyJ0ZW1wVUlEcyI6eyIzM2Fjcm9zcyI6eyJ1aWQiOiIxMTkxNzkxMDk5Nzc2NjEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTo0My4zODg4Nzk5NVoifSwiYWRmIjp7InVpZCI6IjgwNDQ2MDgzMzM3Nzg4MzkwNzgiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMS4wMzMwNTQ3MjdaIn0sImFka2VybmVsIjp7InVpZCI6IkE5MTYzNTAwNzE0OTkyOTMyOTkwIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuMzczMzg1NjYyWiJ9LCJhZGtlcm5lbEFkbiI6eyJ1aWQiOiJBOTE2MzUwMDcxNDk5MjkzMjk5MCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEzLjQzNDkyNTg5NloifSwiYWRtaXhlciI6eyJ1aWQiOiIzNjZhMTdiMTJmMjI0ZDMwOGYzZTNiOGRhOGMzYzhhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjU5MjkxNDgwMVoifSwiYWRueHMiOnsidWlkIjoiNDE5Mjg5ODUzMDE0NTExOTMiLCJleHBpcmVzIjoiMjAyMy0wMS0xOFQwOTo1MzowOC44MjU0NDI2NzZaIn0sImFqYSI6eyJ1aWQiOiJzMnN1aWQ2RGVmMFl0bjJveGQ1aG9zS1AxVmV3IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTMuMjM5MTc2MDU0WiJ9LCJlcGxhbm5pbmciOnsidWlkIjoiQUoxRjBTOE5qdTdTQ0xWOSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjkyOTk2MDQ3M1oifSwiZ2Ftb3NoaSI6eyJ1aWQiOiJndXNyXzM1NmFmOWIxZDhjNjQyYjQ4MmNiYWQyYjdhMjg4MTYxIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuNTI0MTU3MjI1WiJ9LCJncmlkIjp7InVpZCI6IjRmYzM2MjUwLWQ4NTItNDU5Yy04NzcyLTczNTZkZTE3YWI5NyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE0LjY5NjMxNjIyN1oifSwiZ3JvdXBtIjp7InVpZCI6IjdENzVEMjVGLUZBQzktNDQzRC1CMkQxLUIxN0ZFRTExRTAyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjM5LjIyNjIxMjUzMloifSwiaXgiOnsidWlkIjoiWW9ORlNENlc5QkphOEh6eEdtcXlCUUFBXHUwMDI2Mjk3IiwiZXhwaXJlcyI6IjIwMjMtMDUtMzFUMDc6NTM6MzguNTU1ODI3MzU0WiJ9LCJqaXhpZSI6eyJ1aWQiOiI3MzY3MTI1MC1lODgyLTExZWMtYjUzOC0xM2FjYjdhZjBkZTQiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi4xOTEwOTk3MzJaIn0sImxvZ2ljYWQiOnsidWlkIjoiQVZ4OVROQS11c25pa3M4QURzTHpWa3JvaDg4QUFBR0JUREh0UUEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS40NTUxNDk2MTZaIn0sIm1lZGlhbmV0Ijp7InVpZCI6IjI5Nzg0MjM0OTI4OTU0MTAwMDBWMTAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMy42NzIyMTUxMjhaIn0sIm1naWQiOnsidWlkIjoibTU5Z1hyN0xlX1htIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTcuMDk3MDAxNDcxWiJ9LCJuYW5vaW50ZXJhY3RpdmUiOnsidWlkIjoiNmFlYzhjMTAzNzlkY2I3ODQxMmJjODBiNmRkOWM5NzMxNzNhYjdkNzEyZTQzMWE1YTVlYTcwMzRlNTZhNThhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE2LjcxNDgwNzUwNVoifSwib25ldGFnIjp7InVpZCI6IjdPelZoVzFOeC1LOGFVak1HMG52NXVNYm5YNEFHUXZQbnVHcHFrZ3k0ckEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS4xNDE3NDEyNjJaIn0sIm9wZW54Ijp7InVpZCI6IjVkZWNlNjIyLTBhMjMtMGRhYi0zYTI0LTVhNzcwMTBlNDU4MiIsImV4cGlyZXMiOiIyMDIzLTA1LTMxVDA3OjUyOjQ3LjE0MDQxNzM2M1oifSwicHVibWF0aWMiOnsidWlkIjoiN0Q3NUQyNUYtRkFDOS00NDNELUIyRDEtQjE3RkVFMTFFMDI3IiwiZXhwaXJlcyI6IjIwMjItMTAtMzFUMDk6MTQ6MjUuNzM3MjU2ODk5WiJ9LCJyaWNoYXVkaWVuY2UiOnsidWlkIjoiY2I2YzYzMjAtMzNlMi00Nzc0LWIxNjAtMXp6MTY1NDg0MDc0OSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjUyNTA3NDE4WiJ9LCJzbWFydHlhZHMiOnsidWlkIjoiMTJhZjE1ZTQ0ZjAwZDA3NjMwZTc0YzQ5MTU0Y2JmYmE0Zjg0N2U4ZDRhMTU0YzhjM2Q1MWY1OGNmNzJhNDYyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjgyNTAzMTg4NFoifSwic21pbGV3YW50ZWQiOnsidWlkIjoiZGQ5YzNmZTE4N2VmOWIwOWNhYTViNzExNDA0YzI4MzAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNC4yNTU2MDkzNjNaIn0sInN5bmFjb3JtZWRpYSI6eyJ1aWQiOiJHRFBSIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuOTc5NTgzNDM4WiJ9LCJ0cmlwbGVsaWZ0Ijp7InVpZCI6IjcwMjE5NzUwNTQ4MDg4NjUxOTQ2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA4Ljk4OTY3MzU3NFoifSwidmFsdWVpbXByZXNzaW9uIjp7InVpZCI6IjlkMDgxNTVmLWQ5ZmUtNGI1OC04OThlLWUyYzU2MjgyYWIzZSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjA2NzgzOTE2NFoifSwidmlzeCI6eyJ1aWQiOiIyN2UwYWMzYy1iNDZlLTQxYjMtOTkyYy1mOGQyNzE0OTQ5NWUiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi45ODk1MjM1NzNaIn0sInlpZWxkbGFiIjp7InVpZCI6IjY5NzE0ZDlmLWZiMDAtNGE1Zi04MTljLTRiZTE5MTM2YTMyNSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjExLjMwMzAyNjYxNVoifSwieWllbGRtbyI6eyJ1aWQiOiJnOTZjMmY3MTlmMTU1MWIzMWY2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjExMDUyODYwOVoifSwieWllbGRvbmUiOnsidWlkIjoiMmE0MmZiZDMtMmM3MC00ZWI5LWIxYmQtMDQ2OTY2NTBkOTQ4IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuMzE4MzMzOTM5WiJ9LCJ6ZXJvY2xpY2tmcmF1ZCI6eyJ1aWQiOiJiOTk5NThmZS0yYTg3LTJkYTQtOWNjNC05NjFmZDExM2JlY2UiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNS43MTk1OTQ1NjZaIn19LCJiZGF5IjoiMjAyMi0wNS0xN1QwNjo0ODozOC4wMTc5ODgyMDZaIn0=; KRTBCOOKIE_153=1923-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&19420-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&22979-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&23462-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse; KRTBCOOKIE_57=22776-41928985301451193&KRTB&23339-41928985301451193; KRTBCOOKIE_27=16735-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&16736-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&23019-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&23114-uid:3cab6283-4546-4500-a7b6-40ef605fe745; KRTBCOOKIE_18=22947-1978557989514665832; KRTBCOOKIE_466=16530-4fc36250-d852-459c-8772-7356de17ab97; KRTBCOOKIE_391=22924-8044608333778839078&KRTB&23263-8044608333778839078&KRTB&23481-8044608333778839078; KRTBCOOKIE_1310=23431-b81c3g7dr67i&KRTB&23446-b81c3g7dr67i&KRTB&23465-b81c3g7dr67i; KRTBCOOKIE_1290=23368-vkf3yv9lbbl; KRTBCOOKIE_22=14911-4554572065121110164&KRTB&23150-4554572065121110164; KRTBCOOKIE_860=16335-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23334-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23417-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23426-YGAqDU1zUTdjyAFxCoe3kctlNPo; KRTBCOOKIE_904=16787-KwJwE7NkCZClNJRysN2iYg; KRTBCOOKIE_1159=23138-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23328-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23427-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23445-5545f53f3d6e4ec199d8ed627ff026f3; KRTBCOOKIE_32=11175-AQEI_1QecY2ESAIjEW6KAQEBAQE&KRTB&22713-AQEI_1QecY2ESAIjEW6KAQEBAQE&KRTB&22715-AQEI_1QecY2ESAIjEW6KAQEBAQE; SyncRTB3=1685577600%3A35%7C1685491200%3A107_21_71_56_204_247_165_231_233_179_22_209_54_254_238_96_99_220_7_214_13_3_8_234_176_46_5%7C1684886400%3A2_223_15%7C1689465600%3A69%7C1685145600%3A63; KRTBCOOKIE_107=1471-uid:EK38R0PM1NQR0H5&KRTB&23421-uid:EK38R0PM1NQR0H5; KRTBCOOKIE_594=17105-RX-447a6332-530e-456a-97f4-3f0fd1ed48c9-004&KRTB&17107-RX-447a6332-530e-456a-97f4-3f0fd1ed48c9-004; SPugT=1684310122; chkChromeAb67Sec=133; KRTBCOOKIE_699=22727-AAFy2k7FBosAAEasbJoXnw; PugT=1684310473; origin=go-test`)
						return r
					}(),
					Body: []byte(`{"id":"test-case-1","at":1,"bcat":["IAB26-4","IAB26-2","IAB25-6","IAB25-5","IAB25-4","IAB25-3","IAB25-1","IAB25-7","IAB8-18","IAB26-3","IAB26-1","IAB8-5","IAB25-2","IAB11-4"],"tmax":1000,"app":{"publisher":{"name":"New Story Inc.","id":"111","ext":{"installed_sdk":{"id":"MOLOCO_BIDDING","sdk_version":{"major":1,"minor":0,"micro":0},"adapter_version":{"major":1,"minor":0,"micro":0}}}},"paid":0,"name":"DrawHappyAngel","ver":"0.5.4","bundle":"com.newstory.DrawHappyAngel","cat":["IAB9-30"],"id":"1234567","ext":{"orientation":1}},"device":{"ifa":"497a10d6-c4dd-4e04-a986-c32b7180d462","ip":"38.158.207.171","carrier":"MYTEL","language":"en_US","hwv":"ruby","ppi":440,"pxratio":2.75,"devicetype":4,"connectiontype":2,"js":1,"h":2400,"w":1080,"geo":{"type":2,"ipservice":3,"lat":40.7429,"lon":-73.9392,"long":-73.9392,"city":"Queens","country":"USA","region":"ny","dma":"501","metro":"501","zip":"11101","ext":{"org":"Myanmar Broadband Telecom Co.","isp":"Myanmar Broadband Telecom Co."}},"ext":{},"osv":"13.0.0","ua":"Mozilla/5.0 (Linux; Android 13; 22101316C Build/TP1A.220624.014; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.230 Mobile Safari/537.36","make":"xiaomi","model":"22101316c","os":"android"},"imp":[{"id":"1","displaymanager":"applovin_mediation","displaymanagerver":"11.8.2","instl":0,"secure":0,"tagid":"/43743431/DMDemo","bidfloor":0.01,"bidfloorcur":"USD","exp":14400,"banner":{"format":[{"w":728,"h":90},{"w":300,"h":250}],"w":700,"h":900},"video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"rwdd":0}],"user":{"data":[{"id":"1","name":"Publisher Passed","segment":[{"signal":"{\"id\":\"95d6643c-3da6-40a2-b9ca-12279393ffbf\",\"at\":1,\"tmax\":500,\"cur\":[\"USD\"],\"imp\":[{\"id\":\"imp176227948\",\"clickbrowser\":0,\"displaymanager\":\"PubMatic_OpenBid_SDK\",\"displaymanagerver\":\"1.4.0\",\"tagid\":\"/43743431/DMDemo\",\"secure\":0,\"banner\":{\"pos\":7,\"format\":[{\"w\":300,\"h\":250}],\"api\":[5,6,7]},\"instl\":1}],\"app\":{\"paid\":4,\"name\":\"OpenWrapperSample\",\"bundle\":\"com.pubmatic.openbid.app\",\"storeurl\":\"https://itunes.apple.com/us/app/pubmatic-sdk-app/id1175273098\",\"ver\":\"1.0\",\"publisher\":{\"id\":\"111\"}},\"device\":{\"geo\":{\"type\":1,\"lat\":37.421998333333335,\"lon\":-122.08400000000002},\"pxratio\":2.625,\"mccmnc\":\"310-260\",\"lmt\":0,\"ifa\":\"07c387f2-e030-428f-8336-42f682150759\",\"connectiontype\":5,\"carrier\":\"Android\",\"js\":1,\"ua\":\"signal_UA\",\"make\":\"Google\",\"model\":\"AndroidSDKbuiltforx86\",\"os\":\"Android\",\"osv\":\"9\",\"h\":1794,\"w\":1080,\"language\":\"en-US\",\"devicetype\":4,\"ext\":{\"atts\":3}},\"source\":{\"ext\":{\"omidpn\":\"PubMatic\",\"omidpv\":\"1.2.11-Pubmatic\"}},\"user\":{\"data\":[{\"id\":\"1234\"}]},\"ext\":{\"wrapper\":{\"ssauction\":1,\"sumry_disable\":0,\"profileid\":58135,\"versionid\":1,\"clientconfig\":1}}}"}]}],"ext":{"gdpr":0}},"regs":{"coppa":0,"ext":{"gdpr":0}},"source":{"ext":{"schain":{"ver":"1.0","complete":1,"nodes":[{"asi":"applovin.com","sid":"53bf468f18c5a0e2b7d4e3f748c677c1","rid":"494dbe15a3ce08c54f4e456363f35a022247f997","hp":1}]}}},"ext":{"prebid":{"bidderparams":{"pubmatic":{"sendburl":true,"wrapper":{"profileid":222,"versionid":1}}}}}}`),
				},
				setup: func(mme *mock_metrics.MockMetricsEngine) {
					mockFeature.EXPECT().GetApplovinSchainABTestPercentage().Return(10)
					mockEngine.EXPECT().RecordRequestWithSchainABTestEnabled()
					GetRandomNumberIn1To100 = func() int {
						return 5
					}
				},
			},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{
				ModuleContext: hookstage.ModuleContext{
					"rctx": models.RequestCtx{
						PubIDStr:                  "111",
						PubID:                     111,
						ProfileID:                 222,
						DisplayID:                 1,
						DisplayVersionID:          1,
						SSAuction:                 -1,
						ClientConfigFlag:          1,
						Debug:                     true,
						DeviceCtx:                 models.DeviceCtx{IP: "127.0.0.1", UA: "go-test"},
						IsCTVRequest:              false,
						TrackerEndpoint:           "t.pubmatic.com",
						VideoErrorTrackerEndpoint: "t.pubmatic.com/error",
						UidCookie: &http.Cookie{
							Name:  "uids",
							Value: `eyJ0ZW1wVUlEcyI6eyIzM2Fjcm9zcyI6eyJ1aWQiOiIxMTkxNzkxMDk5Nzc2NjEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTo0My4zODg4Nzk5NVoifSwiYWRmIjp7InVpZCI6IjgwNDQ2MDgzMzM3Nzg4MzkwNzgiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMS4wMzMwNTQ3MjdaIn0sImFka2VybmVsIjp7InVpZCI6IkE5MTYzNTAwNzE0OTkyOTMyOTkwIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuMzczMzg1NjYyWiJ9LCJhZGtlcm5lbEFkbiI6eyJ1aWQiOiJBOTE2MzUwMDcxNDk5MjkzMjk5MCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEzLjQzNDkyNTg5NloifSwiYWRtaXhlciI6eyJ1aWQiOiIzNjZhMTdiMTJmMjI0ZDMwOGYzZTNiOGRhOGMzYzhhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjU5MjkxNDgwMVoifSwiYWRueHMiOnsidWlkIjoiNDE5Mjg5ODUzMDE0NTExOTMiLCJleHBpcmVzIjoiMjAyMy0wMS0xOFQwOTo1MzowOC44MjU0NDI2NzZaIn0sImFqYSI6eyJ1aWQiOiJzMnN1aWQ2RGVmMFl0bjJveGQ1aG9zS1AxVmV3IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTMuMjM5MTc2MDU0WiJ9LCJlcGxhbm5pbmciOnsidWlkIjoiQUoxRjBTOE5qdTdTQ0xWOSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjkyOTk2MDQ3M1oifSwiZ2Ftb3NoaSI6eyJ1aWQiOiJndXNyXzM1NmFmOWIxZDhjNjQyYjQ4MmNiYWQyYjdhMjg4MTYxIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuNTI0MTU3MjI1WiJ9LCJncmlkIjp7InVpZCI6IjRmYzM2MjUwLWQ4NTItNDU5Yy04NzcyLTczNTZkZTE3YWI5NyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE0LjY5NjMxNjIyN1oifSwiZ3JvdXBtIjp7InVpZCI6IjdENzVEMjVGLUZBQzktNDQzRC1CMkQxLUIxN0ZFRTExRTAyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjM5LjIyNjIxMjUzMloifSwiaXgiOnsidWlkIjoiWW9ORlNENlc5QkphOEh6eEdtcXlCUUFBXHUwMDI2Mjk3IiwiZXhwaXJlcyI6IjIwMjMtMDUtMzFUMDc6NTM6MzguNTU1ODI3MzU0WiJ9LCJqaXhpZSI6eyJ1aWQiOiI3MzY3MTI1MC1lODgyLTExZWMtYjUzOC0xM2FjYjdhZjBkZTQiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi4xOTEwOTk3MzJaIn0sImxvZ2ljYWQiOnsidWlkIjoiQVZ4OVROQS11c25pa3M4QURzTHpWa3JvaDg4QUFBR0JUREh0UUEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS40NTUxNDk2MTZaIn0sIm1lZGlhbmV0Ijp7InVpZCI6IjI5Nzg0MjM0OTI4OTU0MTAwMDBWMTAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMy42NzIyMTUxMjhaIn0sIm1naWQiOnsidWlkIjoibTU5Z1hyN0xlX1htIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTcuMDk3MDAxNDcxWiJ9LCJuYW5vaW50ZXJhY3RpdmUiOnsidWlkIjoiNmFlYzhjMTAzNzlkY2I3ODQxMmJjODBiNmRkOWM5NzMxNzNhYjdkNzEyZTQzMWE1YTVlYTcwMzRlNTZhNThhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE2LjcxNDgwNzUwNVoifSwib25ldGFnIjp7InVpZCI6IjdPelZoVzFOeC1LOGFVak1HMG52NXVNYm5YNEFHUXZQbnVHcHFrZ3k0ckEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS4xNDE3NDEyNjJaIn0sIm9wZW54Ijp7InVpZCI6IjVkZWNlNjIyLTBhMjMtMGRhYi0zYTI0LTVhNzcwMTBlNDU4MiIsImV4cGlyZXMiOiIyMDIzLTA1LTMxVDA3OjUyOjQ3LjE0MDQxNzM2M1oifSwicHVibWF0aWMiOnsidWlkIjoiN0Q3NUQyNUYtRkFDOS00NDNELUIyRDEtQjE3RkVFMTFFMDI3IiwiZXhwaXJlcyI6IjIwMjItMTAtMzFUMDk6MTQ6MjUuNzM3MjU2ODk5WiJ9LCJyaWNoYXVkaWVuY2UiOnsidWlkIjoiY2I2YzYzMjAtMzNlMi00Nzc0LWIxNjAtMXp6MTY1NDg0MDc0OSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjUyNTA3NDE4WiJ9LCJzbWFydHlhZHMiOnsidWlkIjoiMTJhZjE1ZTQ0ZjAwZDA3NjMwZTc0YzQ5MTU0Y2JmYmE0Zjg0N2U4ZDRhMTU0YzhjM2Q1MWY1OGNmNzJhNDYyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjgyNTAzMTg4NFoifSwic21pbGV3YW50ZWQiOnsidWlkIjoiZGQ5YzNmZTE4N2VmOWIwOWNhYTViNzExNDA0YzI4MzAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNC4yNTU2MDkzNjNaIn0sInN5bmFjb3JtZWRpYSI6eyJ1aWQiOiJHRFBSIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuOTc5NTgzNDM4WiJ9LCJ0cmlwbGVsaWZ0Ijp7InVpZCI6IjcwMjE5NzUwNTQ4MDg4NjUxOTQ2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA4Ljk4OTY3MzU3NFoifSwidmFsdWVpbXByZXNzaW9uIjp7InVpZCI6IjlkMDgxNTVmLWQ5ZmUtNGI1OC04OThlLWUyYzU2MjgyYWIzZSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjA2NzgzOTE2NFoifSwidmlzeCI6eyJ1aWQiOiIyN2UwYWMzYy1iNDZlLTQxYjMtOTkyYy1mOGQyNzE0OTQ5NWUiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi45ODk1MjM1NzNaIn0sInlpZWxkbGFiIjp7InVpZCI6IjY5NzE0ZDlmLWZiMDAtNGE1Zi04MTljLTRiZTE5MTM2YTMyNSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjExLjMwMzAyNjYxNVoifSwieWllbGRtbyI6eyJ1aWQiOiJnOTZjMmY3MTlmMTU1MWIzMWY2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjExMDUyODYwOVoifSwieWllbGRvbmUiOnsidWlkIjoiMmE0MmZiZDMtMmM3MC00ZWI5LWIxYmQtMDQ2OTY2NTBkOTQ4IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuMzE4MzMzOTM5WiJ9LCJ6ZXJvY2xpY2tmcmF1ZCI6eyJ1aWQiOiJiOTk5NThmZS0yYTg3LTJkYTQtOWNjNC05NjFmZDExM2JlY2UiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNS43MTk1OTQ1NjZaIn19LCJiZGF5IjoiMjAyMi0wNS0xN1QwNjo0ODozOC4wMTc5ODgyMDZaIn0=`,
						},
						KADUSERCookie: &http.Cookie{
							Name:  "KADUSERCOOKIE",
							Value: `7D75D25F-FAC9-443D-B2D1-B17FEE11E027`,
						},
						OriginCookie:             "go-test",
						Aliases:                  make(map[string]string),
						ImpBidCtx:                make(map[string]models.ImpCtx),
						PrebidBidderCode:         make(map[string]string),
						BidderResponseTimeMillis: make(map[string]int),
						ProfileIDStr:             "222",
						Endpoint:                 models.EndpointAppLovinMax,
						MetricsEngine:            mockEngine,
						SeatNonBids:              make(map[string][]openrtb_ext.NonBid),
						Method:                   "POST",
						WakandaDebug: &wakanda.Debug{
							Enabled:     true,
							FolderPaths: []string{"DC1__PUB:111__PROF:222"},
							DebugLevel:  2,
							DebugData: wakanda.DebugData{
								HTTPRequest: func() *http.Request {
									r, err := http.NewRequest("POST", "http://localhost/openrtb2/auction?source=owsdk&agent=max&debug=1", nil)
									if err != nil {
										panic(err)
									}
									r.Header.Add("User-Agent", "go-test")
									r.Header.Add("Source_ip", "127.0.0.1")
									r.Header.Add("Cookie", `KADUSERCOOKIE=7D75D25F-FAC9-443D-B2D1-B17FEE11E027; DPSync3=1684886400%3A248%7C1685491200%3A245_226_201; KRTBCOOKIE_80=16514-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&22987-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&23025-CAESEMih0bN7ISRdZT8xX8LXzEw&KRTB&23386-CAESEMih0bN7ISRdZT8xX8LXzEw; KRTBCOOKIE_377=6810-59dc50c9-d658-44ce-b442-5a1f344d97c0&KRTB&22918-59dc50c9-d658-44ce-b442-5a1f344d97c0&KRTB&23031-59dc50c9-d658-44ce-b442-5a1f344d97c0; uids=eyJ0ZW1wVUlEcyI6eyIzM2Fjcm9zcyI6eyJ1aWQiOiIxMTkxNzkxMDk5Nzc2NjEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTo0My4zODg4Nzk5NVoifSwiYWRmIjp7InVpZCI6IjgwNDQ2MDgzMzM3Nzg4MzkwNzgiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMS4wMzMwNTQ3MjdaIn0sImFka2VybmVsIjp7InVpZCI6IkE5MTYzNTAwNzE0OTkyOTMyOTkwIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuMzczMzg1NjYyWiJ9LCJhZGtlcm5lbEFkbiI6eyJ1aWQiOiJBOTE2MzUwMDcxNDk5MjkzMjk5MCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEzLjQzNDkyNTg5NloifSwiYWRtaXhlciI6eyJ1aWQiOiIzNjZhMTdiMTJmMjI0ZDMwOGYzZTNiOGRhOGMzYzhhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjU5MjkxNDgwMVoifSwiYWRueHMiOnsidWlkIjoiNDE5Mjg5ODUzMDE0NTExOTMiLCJleHBpcmVzIjoiMjAyMy0wMS0xOFQwOTo1MzowOC44MjU0NDI2NzZaIn0sImFqYSI6eyJ1aWQiOiJzMnN1aWQ2RGVmMFl0bjJveGQ1aG9zS1AxVmV3IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTMuMjM5MTc2MDU0WiJ9LCJlcGxhbm5pbmciOnsidWlkIjoiQUoxRjBTOE5qdTdTQ0xWOSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjkyOTk2MDQ3M1oifSwiZ2Ftb3NoaSI6eyJ1aWQiOiJndXNyXzM1NmFmOWIxZDhjNjQyYjQ4MmNiYWQyYjdhMjg4MTYxIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuNTI0MTU3MjI1WiJ9LCJncmlkIjp7InVpZCI6IjRmYzM2MjUwLWQ4NTItNDU5Yy04NzcyLTczNTZkZTE3YWI5NyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE0LjY5NjMxNjIyN1oifSwiZ3JvdXBtIjp7InVpZCI6IjdENzVEMjVGLUZBQzktNDQzRC1CMkQxLUIxN0ZFRTExRTAyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjM5LjIyNjIxMjUzMloifSwiaXgiOnsidWlkIjoiWW9ORlNENlc5QkphOEh6eEdtcXlCUUFBXHUwMDI2Mjk3IiwiZXhwaXJlcyI6IjIwMjMtMDUtMzFUMDc6NTM6MzguNTU1ODI3MzU0WiJ9LCJqaXhpZSI6eyJ1aWQiOiI3MzY3MTI1MC1lODgyLTExZWMtYjUzOC0xM2FjYjdhZjBkZTQiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi4xOTEwOTk3MzJaIn0sImxvZ2ljYWQiOnsidWlkIjoiQVZ4OVROQS11c25pa3M4QURzTHpWa3JvaDg4QUFBR0JUREh0UUEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS40NTUxNDk2MTZaIn0sIm1lZGlhbmV0Ijp7InVpZCI6IjI5Nzg0MjM0OTI4OTU0MTAwMDBWMTAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMy42NzIyMTUxMjhaIn0sIm1naWQiOnsidWlkIjoibTU5Z1hyN0xlX1htIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTcuMDk3MDAxNDcxWiJ9LCJuYW5vaW50ZXJhY3RpdmUiOnsidWlkIjoiNmFlYzhjMTAzNzlkY2I3ODQxMmJjODBiNmRkOWM5NzMxNzNhYjdkNzEyZTQzMWE1YTVlYTcwMzRlNTZhNThhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE2LjcxNDgwNzUwNVoifSwib25ldGFnIjp7InVpZCI6IjdPelZoVzFOeC1LOGFVak1HMG52NXVNYm5YNEFHUXZQbnVHcHFrZ3k0ckEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS4xNDE3NDEyNjJaIn0sIm9wZW54Ijp7InVpZCI6IjVkZWNlNjIyLTBhMjMtMGRhYi0zYTI0LTVhNzcwMTBlNDU4MiIsImV4cGlyZXMiOiIyMDIzLTA1LTMxVDA3OjUyOjQ3LjE0MDQxNzM2M1oifSwicHVibWF0aWMiOnsidWlkIjoiN0Q3NUQyNUYtRkFDOS00NDNELUIyRDEtQjE3RkVFMTFFMDI3IiwiZXhwaXJlcyI6IjIwMjItMTAtMzFUMDk6MTQ6MjUuNzM3MjU2ODk5WiJ9LCJyaWNoYXVkaWVuY2UiOnsidWlkIjoiY2I2YzYzMjAtMzNlMi00Nzc0LWIxNjAtMXp6MTY1NDg0MDc0OSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjUyNTA3NDE4WiJ9LCJzbWFydHlhZHMiOnsidWlkIjoiMTJhZjE1ZTQ0ZjAwZDA3NjMwZTc0YzQ5MTU0Y2JmYmE0Zjg0N2U4ZDRhMTU0YzhjM2Q1MWY1OGNmNzJhNDYyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjgyNTAzMTg4NFoifSwic21pbGV3YW50ZWQiOnsidWlkIjoiZGQ5YzNmZTE4N2VmOWIwOWNhYTViNzExNDA0YzI4MzAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNC4yNTU2MDkzNjNaIn0sInN5bmFjb3JtZWRpYSI6eyJ1aWQiOiJHRFBSIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuOTc5NTgzNDM4WiJ9LCJ0cmlwbGVsaWZ0Ijp7InVpZCI6IjcwMjE5NzUwNTQ4MDg4NjUxOTQ2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA4Ljk4OTY3MzU3NFoifSwidmFsdWVpbXByZXNzaW9uIjp7InVpZCI6IjlkMDgxNTVmLWQ5ZmUtNGI1OC04OThlLWUyYzU2MjgyYWIzZSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjA2NzgzOTE2NFoifSwidmlzeCI6eyJ1aWQiOiIyN2UwYWMzYy1iNDZlLTQxYjMtOTkyYy1mOGQyNzE0OTQ5NWUiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi45ODk1MjM1NzNaIn0sInlpZWxkbGFiIjp7InVpZCI6IjY5NzE0ZDlmLWZiMDAtNGE1Zi04MTljLTRiZTE5MTM2YTMyNSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjExLjMwMzAyNjYxNVoifSwieWllbGRtbyI6eyJ1aWQiOiJnOTZjMmY3MTlmMTU1MWIzMWY2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjExMDUyODYwOVoifSwieWllbGRvbmUiOnsidWlkIjoiMmE0MmZiZDMtMmM3MC00ZWI5LWIxYmQtMDQ2OTY2NTBkOTQ4IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuMzE4MzMzOTM5WiJ9LCJ6ZXJvY2xpY2tmcmF1ZCI6eyJ1aWQiOiJiOTk5NThmZS0yYTg3LTJkYTQtOWNjNC05NjFmZDExM2JlY2UiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNS43MTk1OTQ1NjZaIn19LCJiZGF5IjoiMjAyMi0wNS0xN1QwNjo0ODozOC4wMTc5ODgyMDZaIn0=; KRTBCOOKIE_153=1923-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&19420-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&22979-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse&KRTB&23462-LZw-Ny-bMjI2m2Q8IpsrNnnIYzw2yjdnLJsrdYse; KRTBCOOKIE_57=22776-41928985301451193&KRTB&23339-41928985301451193; KRTBCOOKIE_27=16735-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&16736-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&23019-uid:3cab6283-4546-4500-a7b6-40ef605fe745&KRTB&23114-uid:3cab6283-4546-4500-a7b6-40ef605fe745; KRTBCOOKIE_18=22947-1978557989514665832; KRTBCOOKIE_466=16530-4fc36250-d852-459c-8772-7356de17ab97; KRTBCOOKIE_391=22924-8044608333778839078&KRTB&23263-8044608333778839078&KRTB&23481-8044608333778839078; KRTBCOOKIE_1310=23431-b81c3g7dr67i&KRTB&23446-b81c3g7dr67i&KRTB&23465-b81c3g7dr67i; KRTBCOOKIE_1290=23368-vkf3yv9lbbl; KRTBCOOKIE_22=14911-4554572065121110164&KRTB&23150-4554572065121110164; KRTBCOOKIE_860=16335-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23334-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23417-YGAqDU1zUTdjyAFxCoe3kctlNPo&KRTB&23426-YGAqDU1zUTdjyAFxCoe3kctlNPo; KRTBCOOKIE_904=16787-KwJwE7NkCZClNJRysN2iYg; KRTBCOOKIE_1159=23138-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23328-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23427-5545f53f3d6e4ec199d8ed627ff026f3&KRTB&23445-5545f53f3d6e4ec199d8ed627ff026f3; KRTBCOOKIE_32=11175-AQEI_1QecY2ESAIjEW6KAQEBAQE&KRTB&22713-AQEI_1QecY2ESAIjEW6KAQEBAQE&KRTB&22715-AQEI_1QecY2ESAIjEW6KAQEBAQE; SyncRTB3=1685577600%3A35%7C1685491200%3A107_21_71_56_204_247_165_231_233_179_22_209_54_254_238_96_99_220_7_214_13_3_8_234_176_46_5%7C1684886400%3A2_223_15%7C1689465600%3A69%7C1685145600%3A63; KRTBCOOKIE_107=1471-uid:EK38R0PM1NQR0H5&KRTB&23421-uid:EK38R0PM1NQR0H5; KRTBCOOKIE_594=17105-RX-447a6332-530e-456a-97f4-3f0fd1ed48c9-004&KRTB&17107-RX-447a6332-530e-456a-97f4-3f0fd1ed48c9-004; SPugT=1684310122; chkChromeAb67Sec=133; KRTBCOOKIE_699=22727-AAFy2k7FBosAAEasbJoXnw; PugT=1684310473; origin=go-test`)
									return r
								}(),
								HTTPRequestBody: []byte(`{"id":"test-case-1","at":1,"bcat":["IAB26-4","IAB26-2","IAB25-6","IAB25-5","IAB25-4","IAB25-3","IAB25-1","IAB25-7","IAB8-18","IAB26-3","IAB26-1","IAB8-5","IAB25-2","IAB11-4"],"tmax":1000,"app":{"publisher":{"name":"New Story Inc.","id":"111","ext":{"installed_sdk":{"id":"MOLOCO_BIDDING","sdk_version":{"major":1,"minor":0,"micro":0},"adapter_version":{"major":1,"minor":0,"micro":0}}}},"paid":0,"name":"DrawHappyAngel","ver":"0.5.4","bundle":"com.newstory.DrawHappyAngel","cat":["IAB9-30"],"id":"1234567","ext":{"orientation":1}},"device":{"ifa":"497a10d6-c4dd-4e04-a986-c32b7180d462","ip":"38.158.207.171","carrier":"MYTEL","language":"en_US","hwv":"ruby","ppi":440,"pxratio":2.75,"devicetype":4,"connectiontype":2,"js":1,"h":2400,"w":1080,"geo":{"type":2,"ipservice":3,"lat":40.7429,"lon":-73.9392,"long":-73.9392,"city":"Queens","country":"USA","region":"ny","dma":"501","metro":"501","zip":"11101","ext":{"org":"Myanmar Broadband Telecom Co.","isp":"Myanmar Broadband Telecom Co."}},"ext":{},"osv":"13.0.0","ua":"Mozilla/5.0 (Linux; Android 13; 22101316C Build/TP1A.220624.014; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.230 Mobile Safari/537.36","make":"xiaomi","model":"22101316c","os":"android"},"imp":[{"id":"1","displaymanager":"applovin_mediation","displaymanagerver":"11.8.2","instl":0,"secure":0,"tagid":"/43743431/DMDemo","bidfloor":0.01,"bidfloorcur":"USD","exp":14400,"banner":{"format":[{"w":728,"h":90},{"w":300,"h":250}],"w":700,"h":900},"video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"rwdd":0}],"user":{"data":[{"id":"1","name":"Publisher Passed","segment":[{"signal":"{\"id\":\"95d6643c-3da6-40a2-b9ca-12279393ffbf\",\"at\":1,\"tmax\":500,\"cur\":[\"USD\"],\"imp\":[{\"id\":\"imp176227948\",\"clickbrowser\":0,\"displaymanager\":\"PubMatic_OpenBid_SDK\",\"displaymanagerver\":\"1.4.0\",\"tagid\":\"/43743431/DMDemo\",\"secure\":0,\"banner\":{\"pos\":7,\"format\":[{\"w\":300,\"h\":250}],\"api\":[5,6,7]},\"instl\":1}],\"app\":{\"paid\":4,\"name\":\"OpenWrapperSample\",\"bundle\":\"com.pubmatic.openbid.app\",\"storeurl\":\"https://itunes.apple.com/us/app/pubmatic-sdk-app/id1175273098\",\"ver\":\"1.0\",\"publisher\":{\"id\":\"111\"}},\"device\":{\"geo\":{\"type\":1,\"lat\":37.421998333333335,\"lon\":-122.08400000000002},\"pxratio\":2.625,\"mccmnc\":\"310-260\",\"lmt\":0,\"ifa\":\"07c387f2-e030-428f-8336-42f682150759\",\"connectiontype\":5,\"carrier\":\"Android\",\"js\":1,\"ua\":\"signal_UA\",\"make\":\"Google\",\"model\":\"AndroidSDKbuiltforx86\",\"os\":\"Android\",\"osv\":\"9\",\"h\":1794,\"w\":1080,\"language\":\"en-US\",\"devicetype\":4,\"ext\":{\"atts\":3}},\"source\":{\"ext\":{\"omidpn\":\"PubMatic\",\"omidpv\":\"1.2.11-Pubmatic\"}},\"user\":{\"data\":[{\"id\":\"1234\"}]},\"ext\":{\"wrapper\":{\"ssauction\":1,\"sumry_disable\":0,\"profileid\":58135,\"versionid\":1,\"clientconfig\":1}}}"}]}],"ext":{"gdpr":0}},"regs":{"coppa":0,"ext":{"gdpr":0}},"source":{"ext":{"schain":{"ver":"1.0","complete":1,"nodes":[{"asi":"applovin.com","sid":"53bf468f18c5a0e2b7d4e3f748c677c1","rid":"494dbe15a3ce08c54f4e456363f35a022247f997","hp":1}]}}},"ext":{"prebid":{"bidderparams":{"pubmatic":{"sendburl":true,"wrapper":{"profileid":222,"versionid":1}}}}}}`),
							},
						},
						SendBurl:                        true,
						ImpCountingMethodEnabledBidders: make(map[string]struct{}),
					},
				},
			},
			doMutate: true,
			wantErr:  nil,
			wantBody: []byte(`{"id":"test-case-1","imp":[{"id":"1","banner":{"format":[{"w":728,"h":90},{"w":300,"h":250}],"w":700,"h":900,"api":[5,6,7]},"video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"displaymanager":"PubMatic_OpenBid_SDK","displaymanagerver":"1.4.0","tagid":"/43743431/DMDemo","bidfloor":0.01,"bidfloorcur":"USD","clickbrowser":0,"secure":0,"exp":14400}],"app":{"id":"1234567","name":"DrawHappyAngel","bundle":"com.newstory.DrawHappyAngel","storeurl":"https://itunes.apple.com/us/app/pubmatic-sdk-app/id1175273098","cat":["IAB9-30"],"ver":"0.5.4","paid":4,"publisher":{"id":"111","name":"New Story Inc.","ext":{"installed_sdk":{"id":"MOLOCO_BIDDING","sdk_version":{"major":1,"minor":0,"micro":0},"adapter_version":{"major":1,"minor":0,"micro":0}}}},"ext":{"orientation":1}},"device":{"geo":{"lat":40.7429,"lon":-73.9392,"type":2,"ipservice":3,"country":"USA","region":"ny","metro":"501","city":"Queens","zip":"11101","ext":{"org":"Myanmar Broadband Telecom Co.","isp":"Myanmar Broadband Telecom Co."}},"ua":"signal_UA","ip":"38.158.207.171","devicetype":4,"make":"xiaomi","model":"AndroidSDKbuiltforx86","os":"android","osv":"13.0.0","hwv":"ruby","h":2400,"w":1080,"ppi":440,"pxratio":2.75,"js":1,"language":"en_US","carrier":"MYTEL","mccmnc":"310-260","connectiontype":5,"ifa":"497a10d6-c4dd-4e04-a986-c32b7180d462","ext":{"atts":3}},"user":{"data":[{"id":"1234"}],"ext":{"gdpr":0}},"at":1,"tmax":1000,"bcat":["IAB26-4","IAB26-2","IAB25-6","IAB25-5","IAB25-4","IAB25-3","IAB25-1","IAB25-7","IAB8-18","IAB26-3","IAB26-1","IAB8-5","IAB25-2","IAB11-4"],"source":{"ext":{"schain":{"ver":"1.0","complete":1,"nodes":[{"asi":"applovin.com","sid":"53bf468f18c5a0e2b7d4e3f748c677c1","rid":"494dbe15a3ce08c54f4e456363f35a022247f997","hp":1}]},"omidpn":"PubMatic","omidpv":"1.2.11-Pubmatic"}},"regs":{"ext":{"gdpr":0}},"ext":{"prebid":{"bidderparams":{"pubmatic":{"sendburl":true,"wrapper":{"profileid":222,"versionid":1,"clientconfig":1}}}}}}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.setup(mockEngine)
			m := OpenWrap{
				cfg:          tt.fields.cfg,
				cache:        tt.fields.cache,
				metricEngine: mockEngine,
				pubFeatures:  mockFeature,
			}
			got, err := m.handleEntrypointHook(tt.args.in0, tt.args.miCtx, tt.args.payload)
			assert.Equal(t, err, tt.wantErr)

			if tt.want.ModuleContext != nil {
				// validate runtime values individually and reset them
				gotRctx := got.ModuleContext["rctx"].(models.RequestCtx)
				if gotRctx.Endpoint == models.EndpointHybrid || gotRctx.Sshb == "1" {
					assert.Equal(t, got, tt.want)
					return
				}

				assert.NotEmpty(t, gotRctx.StartTime)
				gotRctx.StartTime = 0
				assert.NotEmpty(t, gotRctx.GoogleSDK.StartTime)
				gotRctx.GoogleSDK.StartTime = time.Time{}

				wantRctx := tt.want.ModuleContext["rctx"].(models.RequestCtx)
				if wantRctx.LoggerImpressionID == "" {
					assert.Len(t, gotRctx.LoggerImpressionID, 36)
					gotRctx.LoggerImpressionID = ""
				}
				gotRctx.ParsedUidCookie = nil // ignore parsed cookies
				got.ModuleContext["rctx"] = gotRctx
			}

			if tt.doMutate {
				mutations := got.ChangeSet.Mutations()
				assert.NotEmpty(t, mutations, tt.name)
				for _, mut := range mutations {
					result, err := mut.Apply(tt.args.payload)
					assert.Nil(t, err, tt.name)
					assert.Equal(t, tt.wantBody, result.Body)
				}
			}
			assert.Equal(t, tt.want.ModuleContext, got.ModuleContext)
		})
	}
}

func TestGetRequestWrapper(t *testing.T) {
	type args struct {
		payload  hookstage.EntrypointPayload
		result   hookstage.HookResult[hookstage.EntrypointPayload]
		endpoint string
	}
	tests := []struct {
		name    string
		args    args
		want    models.RequestExtWrapper
		wantErr bool
	}{
		{
			name: "EndpointWebS2S",
			args: args{
				payload: hookstage.EntrypointPayload{
					Body: []byte(`{"ext":{"wrapper":{"profileid":5890,"versionid":1},"prebid":{"bidderparams":{"pubmatic":{"wrapper":{"profileid":100,"versionid":2}}}}}}`),
				},
				result:   hookstage.HookResult[hookstage.EntrypointPayload]{},
				endpoint: models.EndpointWebS2S,
			},
			want: models.RequestExtWrapper{
				SSAuctionFlag: -1,
				ProfileId:     100,
				VersionId:     2,
			},
			wantErr: false,
		},
		{
			name: "EndpointV25",
			args: args{
				payload: hookstage.EntrypointPayload{
					Body: []byte(`{"ext":{"wrapper":{"profileid":5890,"versionid":1},"prebid":{"bidderparams":{"pubmatic":{"wrapper":{"profileid":100,"versionid":2}}}}}}`),
				},
				result:   hookstage.HookResult[hookstage.EntrypointPayload]{},
				endpoint: models.EndpointV25,
			},
			want: models.RequestExtWrapper{
				SSAuctionFlag: -1,
				ProfileId:     5890,
				VersionId:     1,
			},
		},
		{
			name: "EndpointAppLovinMax",
			args: args{
				payload: hookstage.EntrypointPayload{
					Body: []byte(`{"app":{"bundle":"com.pubmatic.openbid.app","name":"Sample","publisher":{"id":"156276"},"id":"13137","storeurl":"https://itunes.apple.com/us/app/pubmatic-sdk-app/id1175273098","ver":"1.0"},"at":1,"device":{"carrier":"MYTEL","connectiontype":2,"devicetype":4,"ext":{"atts":2},"geo":{"city":"Queens","country":"USA","dma":"501","ipservice":3,"lat":40.7429,"lon":-73.9392,"long":-73.9392,"metro":"501","region":"ny","type":2,"zip":"11101"},"h":2400,"hwv":"ruby","ifa":"497a10d6-c4dd-4e04-a986-c32b7180d462","ip":"38.158.207.171","js":1,"language":"en_US","make":"xiaomi","model":"22101316c","os":"android","osv":"13.0.0","ppi":440,"pxratio":2.75,"ua":"Mozilla/5.0 (Linux; Android 13; 22101316C Build/TP1A.220624.014; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.230 Mobile Safari/537.36","w":1080},"ext":{"prebid":{"bidderparams":{"pubmatic":{"wrapper":{"profileid":13137}}}}},"id":"95d6643c-3da6-40a2-b9ca-12279393ffbf","imp":[{"banner":{"format":[{"w":728,"h":90},{"w":320,"h":50}],"api":[5,2],"w":700,"h":900},"clickbrowser":0,"displaymanager":"OpenBid_SDK","displaymanagerver":"1.4.0","ext":{"reward":1},"id":"imp176227948","secure":0,"tagid":"OpenWrapBidderBannerAdUnit"}],"regs":{"ext":{"gdpr":0}},"source":{"ext":{"omidpn":"PubMatic","omidpv":"1.2.11-Pubmatic"}},"user":{"data":[{"id":"1","name":"Publisher Passed","segment":[{"signal":"{\"id\":\"95d6643c-3da6-40a2-b9ca-12279393ffbf\",\"at\":1,\"tmax\":500,\"cur\":[\"USD\"],\"imp\":[{\"id\":\"imp176227948\",\"clickbrowser\":0,\"displaymanager\":\"PubMatic_OpenBid_SDK\",\"displaymanagerver\":\"1.4.0\",\"tagid\":\"/43743431/DMDemo\",\"secure\":0,\"banner\":{\"pos\":7,\"format\":[{\"w\":300,\"h\":250}],\"api\":[5,6,7]},\"instl\":1}],\"app\":{\"name\":\"OpenWrapperSample\",\"bundle\":\"com.pubmatic.openbid.app\",\"storeurl\":\"https://itunes.apple.com/us/app/pubmatic-sdk-app/id1175273098?appnexus_banner_fixedbid=1&fixedbid=1\",\"ver\":\"1.0\",\"publisher\":{\"id\":\"5890\"}},\"device\":{\"ext\":{\"atts\":0},\"geo\":{\"type\":1,\"lat\":37.421998333333335,\"lon\":-122.08400000000002},\"pxratio\":2.625,\"mccmnc\":\"310-260\",\"lmt\":0,\"ifa\":\"07c387f2-e030-428f-8336-42f682150759\",\"connectiontype\":6,\"carrier\":\"Android\",\"js\":1,\"ua\":\"Mozilla/5.0(Linux;Android9;AndroidSDKbuiltforx86Build/PSR1.180720.075;wv)AppleWebKit/537.36(KHTML,likeGecko)Version/4.0Chrome/69.0.3497.100MobileSafari/537.36\",\"make\":\"Google\",\"model\":\"AndroidSDKbuiltforx86\",\"os\":\"Android\",\"osv\":\"9\",\"h\":1794,\"w\":1080,\"language\":\"en-US\",\"devicetype\":4},\"source\":{\"ext\":{\"omidpn\":\"PubMatic\",\"omidpv\":\"1.2.11-Pubmatic\"}},\"user\":{},\"ext\":{\"wrapper\":{\"ssauction\":0,\"sumry_disable\":0,\"profileid\":58135,\"versionid\":1,\"clientconfig\":1}}}"}]}],"ext":{"consent":"CP2KIMAP2KIgAEPgABBYJGNX_H__bX9j-Xr3"}}}`),
				},
				result:   hookstage.HookResult[hookstage.EntrypointPayload]{},
				endpoint: models.EndpointAppLovinMax,
			},
			want: models.RequestExtWrapper{
				SSAuctionFlag: -1,
				ProfileId:     13137,
			},
		},
		{
			name: "EndpointAppLovinMax give preferance to profileid in wrapper",
			args: args{
				payload: hookstage.EntrypointPayload{
					Body: []byte(`{"app":{"bundle":"com.pubmatic.openbid.app","name":"Sample","publisher":{"id":"156276"},"id":"1234","storeurl":"https://itunes.apple.com/us/app/pubmatic-sdk-app/id1175273098","ver":"1.0"},"at":1,"device":{"carrier":"MYTEL","connectiontype":2,"devicetype":4,"ext":{"atts":2},"geo":{"city":"Queens","country":"USA","dma":"501","ipservice":3,"lat":40.7429,"lon":-73.9392,"long":-73.9392,"metro":"501","region":"ny","type":2,"zip":"11101"},"h":2400,"hwv":"ruby","ifa":"497a10d6-c4dd-4e04-a986-c32b7180d462","ip":"38.158.207.171","js":1,"language":"en_US","make":"xiaomi","model":"22101316c","os":"android","osv":"13.0.0","ppi":440,"pxratio":2.75,"ua":"Mozilla/5.0 (Linux; Android 13; 22101316C Build/TP1A.220624.014; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.230 Mobile Safari/537.36","w":1080},"ext":{"prebid":{"bidderparams":{"pubmatic":{"wrapper":{"profileid":5890}}}}},"id":"95d6643c-3da6-40a2-b9ca-12279393ffbf","imp":[{"banner":{"format":[{"w":728,"h":90},{"w":320,"h":50}],"api":[5,2],"w":700,"h":900},"clickbrowser":0,"displaymanager":"OpenBid_SDK","displaymanagerver":"1.4.0","ext":{"reward":1},"id":"imp176227948","secure":0,"tagid":"OpenWrapBidderBannerAdUnit"}],"regs":{"ext":{"gdpr":0}},"source":{"ext":{"omidpn":"PubMatic","omidpv":"1.2.11-Pubmatic"}},"user":{"data":[{"id":"1","name":"Publisher Passed","segment":[{"signal":"{\"id\":\"95d6643c-3da6-40a2-b9ca-12279393ffbf\",\"at\":1,\"tmax\":500,\"cur\":[\"USD\"],\"imp\":[{\"id\":\"imp176227948\",\"clickbrowser\":0,\"displaymanager\":\"PubMatic_OpenBid_SDK\",\"displaymanagerver\":\"1.4.0\",\"tagid\":\"/43743431/DMDemo\",\"secure\":0,\"banner\":{\"pos\":7,\"format\":[{\"w\":300,\"h\":250}],\"api\":[5,6,7]},\"instl\":1}],\"app\":{\"name\":\"OpenWrapperSample\",\"bundle\":\"com.pubmatic.openbid.app\",\"storeurl\":\"https://itunes.apple.com/us/app/pubmatic-sdk-app/id1175273098?appnexus_banner_fixedbid=1&fixedbid=1\",\"ver\":\"1.0\",\"publisher\":{\"id\":\"5890\"}},\"device\":{\"ext\":{\"atts\":0},\"geo\":{\"type\":1,\"lat\":37.421998333333335,\"lon\":-122.08400000000002},\"pxratio\":2.625,\"mccmnc\":\"310-260\",\"lmt\":0,\"ifa\":\"07c387f2-e030-428f-8336-42f682150759\",\"connectiontype\":6,\"carrier\":\"Android\",\"js\":1,\"ua\":\"Mozilla/5.0(Linux;Android9;AndroidSDKbuiltforx86Build/PSR1.180720.075;wv)AppleWebKit/537.36(KHTML,likeGecko)Version/4.0Chrome/69.0.3497.100MobileSafari/537.36\",\"make\":\"Google\",\"model\":\"AndroidSDKbuiltforx86\",\"os\":\"Android\",\"osv\":\"9\",\"h\":1794,\"w\":1080,\"language\":\"en-US\",\"devicetype\":4},\"source\":{\"ext\":{\"omidpn\":\"PubMatic\",\"omidpv\":\"1.2.11-Pubmatic\"}},\"user\":{},\"ext\":{\"wrapper\":{\"ssauction\":0,\"sumry_disable\":0,\"profileid\":58135,\"versionid\":1,\"clientconfig\":1}}}"}]}],"ext":{"consent":"CP2KIMAP2KIgAEPgABBYJGNX_H__bX9j-Xr3"}}}`),
				},
				result:   hookstage.HookResult[hookstage.EntrypointPayload]{},
				endpoint: models.EndpointAppLovinMax,
			},
			want: models.RequestExtWrapper{
				SSAuctionFlag: -1,
				ProfileId:     5890,
			},
		},
		//TODO: Add test cases for other endpoints after migration(AMP, Video, VAST, JSON, InappVideo)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRequestWrapper(tt.args.payload, tt.args.result, tt.args.endpoint)
			assert.Equal(t, err != nil, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetEndpoint(t *testing.T) {
	type args struct {
		path   string
		source string
		agent  string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "EndpointAuction Activation",
			args: args{
				path:   hookexecution.EndpointAuction,
				source: "pbjs",
			},
			want: models.EndpointWebS2S,
		},
		{
			name: "EndpointAuction inapp",
			args: args{
				path:   hookexecution.EndpointAuction,
				source: "owsdk",
			},
			want: models.EndpointV25,
		},
		{
			name: "EndpointAuction default",
			args: args{
				path: hookexecution.EndpointAuction,
			},
			want: models.EndpointHybrid,
		},
		{
			name: "OpenWrapAuction",
			args: args{
				path: OpenWrapAuction,
			},
			want: models.EndpointHybrid,
		},
		{
			name: "OpenWrapV25",
			args: args{
				path: OpenWrapV25,
			},
			want: models.EndpointV25,
		},
		{
			name: "OpenWrapV25Video",
			args: args{
				path: OpenWrapV25Video,
			},
			want: models.EndpintInappVideo,
		},
		{
			name: "OpenWrapAmp",
			args: args{
				path: OpenWrapAmp,
			},
			want: models.EndpointAMP,
		},
		{
			name: "OpenWrapOpenRTBVideo",
			args: args{
				path: OpenWrapOpenRTBVideo,
			},
			want: models.EndpointORTB,
		},
		{
			name: "OpenWrapVAST",
			args: args{
				path: OpenWrapVAST,
			},
			want: models.EndpointVAST,
		},
		{
			name: "OpenWrapJSON",
			args: args{
				path: OpenWrapJSON,
			},
			want: models.EndpointJson,
		},
		{
			name: "ApplovinMax",
			args: args{
				path:   hookexecution.EndpointAuction,
				source: "owsdk",
				agent:  models.AppLovinMaxAgent,
			},
			want: models.EndpointAppLovinMax,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetEndpoint(tt.args.path, tt.args.source, tt.args.agent)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetSendBurl(t *testing.T) {
	type args struct {
		request  []byte
		endpoint string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "AppLovinMax endpoint should always return true",
			args: args{
				request:  []byte(`{}`),
				endpoint: models.EndpointAppLovinMax,
			},
			want: true,
		},
		{
			name: "GoogleSDK endpoint should always return true",
			args: args{
				request:  []byte(`{}`),
				endpoint: models.EndpointGoogleSDK,
			},
			want: true,
		},
		{
			name: "UnityLevelPlay endpoint should always return true",
			args: args{
				request:  []byte(`{}`),
				endpoint: models.EndpointUnityLevelPlay,
			},
			want: true,
		},
		{
			name: "request with sendburl true for other endpoint",
			args: args{
				request:  []byte(`{"ext":{"prebid":{"bidderparams":{"pubmatic":{"sendburl":true,"wrapper":{"profileid":14052,"sumry_disable":1,"clientconfig":1}}}}}}`),
				endpoint: models.EndpointWebS2S,
			},
			want: true,
		},
		{
			name: "request with sendburl false for other endpoint",
			args: args{
				request:  []byte(`{"ext":{"prebid":{"bidderparams":{"pubmatic":{"sendburl":false,"wrapper":{"profileid":14052,"sumry_disable":1,"clientconfig":1}}}}}}`),
				endpoint: models.EndpointWebS2S,
			},
			want: false,
		},
		{
			name: "request with no sendburl key for other endpoint",
			args: args{
				request:  []byte(`{"ext":{"prebid":{"bidderparams":{"pubmatic":{"wrapper":{"profileid":14052,"sumry_disable":1,"clientconfig":1}}}}}}`),
				endpoint: models.EndpointWebS2S,
			},
			want: false,
		},
		{
			name: "no ext object in request for other endpoint",
			args: args{
				request:  []byte(``),
				endpoint: models.EndpointWebS2S,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getSendBurl(tt.args.request, tt.args.endpoint)
			assert.Equal(t, tt.want, got)
		})
	}
}
