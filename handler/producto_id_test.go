package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Un ID fuera de int32 no debe truncarse y consultar o borrar otro producto.
func TestProductoRechazaIDFueraDeRango(t *testing.T) {
	h := NewProductoHandler(nil)
	mux := http.NewServeMux()
	mux.HandleFunc("/productos/{id}", h.Elemento)
	for _, method := range []string{"GET", "PUT", "DELETE"} {
		for _, id := range []string{"0", "-1", "abc", "2147483648", "4294967297"} {
			t.Run(method+"/"+id, func(t *testing.T) {
				res := httptest.NewRecorder()
				mux.ServeHTTP(res, httptest.NewRequest(method, "/productos/"+id, nil))
				if res.Code != http.StatusBadRequest {
					t.Fatalf("status=%d; se esperaba 400", res.Code)
				}
			})
		}
	}
}
