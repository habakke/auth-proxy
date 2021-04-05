package healthz

import "net/http"

type Healthz struct {
}

func Handler() *Healthz {
	return &Healthz{}
}

func (h *Healthz) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(http.StatusOK)
}
