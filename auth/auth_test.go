package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "testpassword123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == "" {
		t.Error("HashPassword returned empty string")
	}

	if hash == password {
		t.Error("Hash should not equal the original password")
	}
}

func TestCheckPasswordHash_Correct(t *testing.T) {
	password := "mypassword"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	result := CheckPasswordHash(password, hash)
	if !result {
		t.Error("CheckPasswordHash should return true for correct password")
	}
}

func TestCheckPasswordHash_Wrong(t *testing.T) {
	password := "correctpassword"
	wrongPassword := "wrongpassword"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	result := CheckPasswordHash(wrongPassword, hash)
	if result {
		t.Error("CheckPasswordHash should return false for wrong password")
	}
}

func TestHashPassword_DifferentHashes(t *testing.T) {
	password := "samepassword"

	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("First HashPassword failed: %v", err)
	}

	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Second HashPassword failed: %v", err)
	}

	if hash1 == hash2 {
		t.Error("Same password should produce different hashes due to salt")
	}

	if !CheckPasswordHash(password, hash1) {
		t.Error("First hash should verify correctly")
	}
	if !CheckPasswordHash(password, hash2) {
		t.Error("Second hash should verify correctly")
	}
}
