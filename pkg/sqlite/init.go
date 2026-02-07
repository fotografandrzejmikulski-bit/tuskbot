package sqlite

/*
#cgo CFLAGS: -Wno-deprecated-declarations
#include <sqlite3.h>

// Forward declaration of the entry point in libsqlite_vec.a
int sqlite3_vec_init(sqlite3*, char**, const sqlite3_api_routines*);

// Helper to register the extension globally
void register_vec_extension(void) {
    sqlite3_auto_extension((void(*)(void))sqlite3_vec_init);
}
*/
import "C"
import (
	"database/sql"

	"github.com/mattn/go-sqlite3"
)

func init() {
	C.register_vec_extension()

	sql.Register("sqlite3_vec", &sqlite3.SQLiteDriver{
		ConnectHook: func(conn *sqlite3.SQLiteConn) error {
			// No need to register here anymore
			return nil
		},
	})
}
