package main


import (
   "fmt"
   "net/http"
)


func main() {
   //defino el contenido del HTML
   htmlContent := `<!DOCTYPE html>
   <html> 
         <head> <title>Hola mundo </title> </head>              
         <body> <h1>!SERVIDOR FUNCIONANDO</h1> </body>
   </html>

   `

   //defino el manejador "HANDLER" de la ruta "/"
   http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html, charset=utf-8")
    fmt.Fprint(w, htmlContent)//escribo el contenido del HTML en la rta
   })
   //Defino el puerto en el que se ejecutará el servidor
   port := ":8080"
   fmt.Printf("Servidor escuchando en http://localhost%s\n", port)


   err := http.ListenAndServe(port, nil)
   if err != nil {
       fmt.Printf("Error al iniciar el servidor: %s\n", err)
   }


//defino el manejador "HANDLER" de la ruta "/"
   http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
       if r.URL.Path != "/" {
           http.NotFound(w, r) 
           return
       }
       w.Header().Set("Content-Type", "text/html, charset=utf-8")
       fmt.Fprint(w, htmlContent) 
   })

}