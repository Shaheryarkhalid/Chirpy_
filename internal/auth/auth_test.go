package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)


func TestValidateToken(t *testing.T){
	var userId uuid.UUID = uuid.New()
	var tokenSecret string = "test_ligma_balls"
	tokenString , err := MakeJWT(userId, tokenSecret, time.Duration(time.Minute * 20))
	if err != nil {
		t.Errorf("Error: Trying to create token:\n%v\n", err)
		t.Fail()
		return
	}
	validatedUserId, err:= ValidateJWT(tokenString, tokenSecret)
	if err != nil {
		t.Errorf("Error: Trying to vlaidate token:\n%v\n", err)
		t.Fail()
		return
	}
	if validatedUserId != userId{
		t.Error("returned userId does not match with the actual userId")
		t.Fail()
		return
	}
}

func TestMakeRefreshToken(t *testing.T){
	token, err := MakeRefreshToken()
	if err != nil || token == "" {
		t.Error(err)
		t.Fail()
	}
}
