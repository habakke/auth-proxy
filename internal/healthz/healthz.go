package healthz

import "net/http"

type Healthz struct {
}

func Handler() *Healthz {
	return &Healthz{}
}

func (h *Healthz) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	_, _ = res.Write([]byte("OK"))
	res.WriteHeader(http.StatusOK)
}
