package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log"
)

func main() {
	// Password del primer admin: Admin123!
	password := "Admin123!"
	
	// Generar hash con bcrypt (cost 12 para producción)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Println("============================================")
	fmt.Println("PRIMER USUARIO ADMINISTRADOR")
	fmt.Println("============================================")
	fmt.Println("Email:    admin@codeplex.com")
	fmt.Println("Password: Admin123!")
	fmt.Println("============================================")
	fmt.Println("\nSQL para insertar el usuario:\n")
	fmt.Printf(`INSERT INTO usuarios_admin (
    email,
    nombre_completo,
    password_hash,
    rol,
    debe_cambiar_password,
    activo,
    creado_por
) VALUES (
    'admin@codeplex.com',
    'Administrador Principal',
    '%s',
    'ADMIN',
    false,
    true,
    NULL
);
`, string(hash))
	fmt.Println("\n============================================")
	fmt.Println("IMPORTANTE:")
	fmt.Println("1. Guarda este SQL en un lugar seguro")
	fmt.Println("2. Ejecuta el schema_usuarios_admin.sql primero")
	fmt.Println("3. Luego ejecuta este INSERT")
	fmt.Println("4. Cambia la contraseña después del primer login")
	fmt.Println("============================================")
}