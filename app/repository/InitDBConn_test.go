package repository

import (
	"context"
	"strings"
	"testing"
)

func TestInitDBConn_InvalidURL(t *testing.T) {
	ctx := context.Background()
	invalidURL := "invalidUrl"
	_, err := InitDBConn(ctx, invalidURL)
	if err == nil {
		t.Fatal("Expected error but got nil")
	}
	expectedErrMsg := "failed to parse pg config"
	if !strings.Contains(err.Error(), expectedErrMsg) {
		t.Errorf("Expected error message to contain '%s', got '%s'", expectedErrMsg, err.Error())
	}
}
func TestInitDBConn_UnreachableServer(t *testing.T) {
	ctx := context.Background()
	dbURL := "postgres://postgres:193566@db:9753/AuthDB?sslmode=disable"
	_, err := InitDBConn(ctx, dbURL)
	if err == nil {
		t.Fatalf("Expected error but got nil")
	}
	expectedErrMsg := "failed to connect config"
	if !strings.Contains(err.Error(), expectedErrMsg) {
		t.Errorf("Expected error message to contain '%s', got '%s'", expectedErrMsg, err.Error())
	}
}

func TestInitDBConn_InsufficientPrivileges(t *testing.T) {
	ctx := context.Background()
	dbURL := "postgres://postgres:WRONGPASSWORD@db:5432/AuthDB?sslmode=disable"
	_, err := InitDBConn(ctx, dbURL)
	if err == nil {
		t.Errorf("Expected error but got nil")
	}
	expectedErrMsg := "failed to connect config"
	if !strings.Contains(err.Error(), expectedErrMsg) {
		t.Errorf("Expected error message to contain '%s', got '%s'", expectedErrMsg, err.Error())
	}
}
