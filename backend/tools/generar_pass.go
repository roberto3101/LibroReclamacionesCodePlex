package main

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	// 1. Define la contraseña segura
	password := "Admin123!"

	// 2. Genera el hash (Coste 12 para producción)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		log.Fatal(err)
	}

	// 3. Imprime el SQL listo para copiar y pegar
	fmt.Println("\n--- COPIA DESDE AQUI ---")
	fmt.Printf(`
DELETE FROM usuarios_admin WHERE email = 'admin@codeplex.com';

INSERT INTO usuarios_admin (
    email,
    nombre_completo,
    password_hash,
    rol,
    debe_cambiar_password,
    activo
) VALUES (
    'admin@codeplex.com',
    'Administrador Principal',
    '%s',
    'ADMIN',
    false,
    true
);
`, string(hash))
	fmt.Println("--- HASTA AQUI ---\n")
}