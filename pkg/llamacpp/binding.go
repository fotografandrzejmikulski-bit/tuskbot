package llamacpp

/*
#include "llama.h"
*/
import "C"
import "fmt"

func TestLink() {
	fmt.Println("[CGO] Calling llama_backend_init()...")

	C.llama_backend_init()

	fmt.Println("[CGO] Backend initialized!")

	C.llama_backend_free()
	fmt.Println("[CGO] Backend freed.")
}
