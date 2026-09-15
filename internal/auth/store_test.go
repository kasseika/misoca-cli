// store_test.go は認証情報（OAuth2トークン）のファイル永続化を検証します。
package auth

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// TestStore_SaveAndLoad は、保存した認証情報をそのまま読み込めることを検証します。
func TestStore_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(filepath.Join(dir, "credentials.json"))

	want := &Credentials{
		ClientID:     "client-id-山田商事",
		ClientSecret: "client-secret-abc",
		AccessToken:  "access-token-abc",
		RefreshToken: "refresh-token-xyz",
		Expiry:       time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC),
		Email:        "yamada@example.co.jp",
	}

	if err := store.Save("default", want); err != nil {
		t.Fatalf("Save が失敗しました: %v", err)
	}

	got, err := store.Load("default")
	if err != nil {
		t.Fatalf("Load が失敗しました: %v", err)
	}
	if *got != *want {
		t.Errorf("Load結果 = %+v, 期待値 = %+v", got, want)
	}
}

// TestStore_Load_NotFound は、未保存のプロファイルを読み込もうとした場合に
// ErrProfileNotFound が返されることを検証します。
func TestStore_Load_NotFound(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(filepath.Join(dir, "credentials.json"))

	_, err := store.Load("default")
	if !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("ErrProfileNotFound を期待しましたが: %v", err)
	}
}

// TestStore_Save_FilePermissions は、保存された認証情報ファイルとその
// ディレクトリが他ユーザーから読めないパーミッション（0600/0700）で
// 作成されることを検証します（Windowsではファイルパーミッションの意味が
// 異なるためスキップします）。
func TestStore_Save_FilePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windowsではposixパーミッションを検証しない")
	}

	dir := t.TempDir()
	credsDir := filepath.Join(dir, "misoca")
	credsPath := filepath.Join(credsDir, "credentials.json")
	store := NewStore(credsPath)

	if err := store.Save("default", &Credentials{AccessToken: "t"}); err != nil {
		t.Fatalf("Save が失敗しました: %v", err)
	}

	fileInfo, err := os.Stat(credsPath)
	if err != nil {
		t.Fatalf("ファイル情報の取得に失敗しました: %v", err)
	}
	if perm := fileInfo.Mode().Perm(); perm != 0o600 {
		t.Errorf("ファイルパーミッション = %o, 期待値 = %o", perm, 0o600)
	}

	dirInfo, err := os.Stat(credsDir)
	if err != nil {
		t.Fatalf("ディレクトリ情報の取得に失敗しました: %v", err)
	}
	if perm := dirInfo.Mode().Perm(); perm != 0o700 {
		t.Errorf("ディレクトリパーミッション = %o, 期待値 = %o", perm, 0o700)
	}
}

// TestStore_Delete は、保存済みプロファイルを削除できることを検証します。
func TestStore_Delete(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(filepath.Join(dir, "credentials.json"))

	if err := store.Save("default", &Credentials{AccessToken: "t"}); err != nil {
		t.Fatalf("Save が失敗しました: %v", err)
	}
	if err := store.Delete("default"); err != nil {
		t.Fatalf("Delete が失敗しました: %v", err)
	}

	_, err := store.Load("default")
	if !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("削除後は ErrProfileNotFound を期待しましたが: %v", err)
	}
}

// TestStore_MultipleProfiles は、複数プロファイルを独立して保存・読み込み
// できることを検証します（個人アカウント・法人アカウントの併用を想定）。
func TestStore_MultipleProfiles(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(filepath.Join(dir, "credentials.json"))

	personal := &Credentials{AccessToken: "personal-token", Email: "individual@example.co.jp"}
	work := &Credentials{AccessToken: "work-token", Email: "work@example.co.jp"}

	if err := store.Save("personal", personal); err != nil {
		t.Fatalf("Save(personal) が失敗しました: %v", err)
	}
	if err := store.Save("work", work); err != nil {
		t.Fatalf("Save(work) が失敗しました: %v", err)
	}

	gotPersonal, err := store.Load("personal")
	if err != nil {
		t.Fatalf("Load(personal) が失敗しました: %v", err)
	}
	if *gotPersonal != *personal {
		t.Errorf("personal = %+v, 期待値 = %+v", gotPersonal, personal)
	}

	gotWork, err := store.Load("work")
	if err != nil {
		t.Fatalf("Load(work) が失敗しました: %v", err)
	}
	if *gotWork != *work {
		t.Errorf("work = %+v, 期待値 = %+v", gotWork, work)
	}
}
