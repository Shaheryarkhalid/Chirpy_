package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/Shaheryarkhalid/Chirpy/internal/auth"
	"github.com/Shaheryarkhalid/Chirpy/internal/database"
	"github.com/google/uuid"
)
type Handlers struct{
	apiConfig *ApiConfig
}

func NewHandler(cnfg *ApiConfig)*Handlers{
	return &Handlers{apiConfig: cnfg}
}

func (h Handlers)handlerMetrics(w http.ResponseWriter, r *http.Request){
	returnedString := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", h.apiConfig.fileserverHits.Load())
	respondWithJson(w, 200, &returnedString) 
}

func (h Handlers)handlerReset(w http.ResponseWriter, r *http.Request){
	h.apiConfig.fileserverHits.Store(0)
	if h.apiConfig.Platform != "dev"{
		respondWithError(w, 403, "403 Forbiden")
		return
	}
	err := h.apiConfig.db.ClearUsers(r.Context())
	if err != nil {
		fmt.Println("Error: Trying to clear Users from db.")
		fmt.Println(err)
		respondWithError(w, 500,"500 Internal Server Error" )
		return 
	}
	returnedString := fmt.Sprintf("Hits: %v", h.apiConfig.fileserverHits.Load())
	respondWithJson(w, 200, &returnedString )
}

func (h Handlers) handlerHealthz(w http.ResponseWriter, _ *http.Request){
	returnedString := "OK"
	respondWithJson(w, 200, &returnedString)
}

func (_ Handlers) handlerCrash(_ http.ResponseWriter,_ *http.Request){
	var p *int
	fmt.Println(*p)
}

func (h Handlers)handlerCreateChirp(w http.ResponseWriter,r *http.Request){
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "Invalid auth token")
		return
	}
	userId, err:= auth.ValidateJWT(tokenString, h.apiConfig.tokenSecret)
	if err != nil {
		respondWithError(w, 401, "Invalid auth token")
		return
	}
	type requestBody struct {
		Body string `json:"body"`
	}
	defer r.Body.Close()
	body := requestBody{}
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		respondWithError(w, 400, "Invalid Chip")
	}
	if len(body.Body) <= 140 {
		body.Body = cleanBody(body.Body)
		createChirpParams := database.CreateChirpParams{Body: body.Body, UserID: userId}
		createdChirp , err := h.apiConfig.db.CreateChirp(r.Context(), createChirpParams)
		if err != nil {
			fmt.Println("Error: trying to add Chirp to the database.")
			fmt.Println(err)
			respondWithError(w, 500, "500: Internal server error.")
			return 
		}
		mappedChirp := Chirp(createdChirp)
		respondWithJson(w, 201, &mappedChirp)
		return 
	}
	respondWithError(w, 400, "Invalid Chirp, must be less than 140 characters.")
}

func (h Handlers) handlerCreateUser(w http.ResponseWriter,r *http.Request){
	reqBody := struct{
		Email string `json:"email"`
		Password string `json:"password"`
	}{}

	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		respondWithError(w, 400, "Invalid data. \"email\" and \"password\" must be present the the included body.")
		return
	}
	type User struct {
		ID        uuid.UUID `json:"id"`
		Email     string `json:"email"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		IsChirpyRed bool `json:"is_chirpy_red"`
	}
	hashedPassword, err := auth.HashPassword(reqBody.Password)
	if err != nil {
		respondWithError(w, 400, "Unable to hash your password. Please try again")
		return
	}
	createUserParams := database.CreateUserParams{
		Email: reqBody.Email,
		HashedPassword: hashedPassword,
		
	}
	createdUser, err := h.apiConfig.db.CreateUser(r.Context(), createUserParams)
	if err != nil {
		fmt.Printf("Error: Trying to add user to database. \n%v\n", err)
		respondWithError(w, 500, "500 Internal Server Error: Trying to add user to database.")
		return
	}
	mappedUser := User(createdUser)
	respondWithJson(w, 201, &mappedUser)
}

func (h Handlers) handlerLogin(w http.ResponseWriter,r *http.Request){
	reqBody := struct {
		Email string `json:"email"` 
		Password string `json:"password"` 
	}{}
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		respondWithError(w, 400, "Request body must include \"email\": string and \"password\": string")
		return
	}
	accessTokenExpiresIn := time.Duration(time.Hour *1)
	user, err := h.apiConfig.db.GetUser(r.Context(), reqBody.Email)
	if err != nil {
		fmt.Println(err)
		respondWithError(w, 401, "Incorrect email or password")
		return
	}
	ok, err := auth.CheckPasswordHash(reqBody.Password, user.HashedPassword)
	if err != nil  || !ok{
		fmt.Println(err)
		respondWithError(w, 401, "Incorrect email or password")
		return 
	}
	accessToken, err := auth.MakeJWT(user.ID, h.apiConfig.tokenSecret, accessTokenExpiresIn)
	if err != nil {
		fmt.Println("Error: Trying to create access token for given user.")
		fmt.Println(err)
	}
	rToken, err := auth.MakeRefreshToken()
	if err != nil {
		fmt.Println("Error: Trying to create refresh token for given user.")
		fmt.Println(err)
	}
	createRefreshTokenParams := database.CreateRefreshTokenParams{
		Token: rToken,
		UserID: user.ID,
		ExpiresAt: time.Now().Add(time.Duration(time.Hour * (24 * 60))),
	}
	refreshToken , err:= h.apiConfig.db.CreateRefreshToken(r.Context(), createRefreshTokenParams)
	if err != nil {
		fmt.Println("Error: Trying to save refresh token in database.")
		fmt.Println(err)
	}
	returnedUser := struct{ 
		ID uuid.UUID `json:"id"`
		Email string `json:"email"`
		IsChirpyRed bool `json:"is_chirpy_red"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Token string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}{
		ID: user.ID,
		Email: user.Email,
		IsChirpyRed: user.IsChirpyRed,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Token: accessToken,
		RefreshToken: refreshToken.Token,
	}
	respondWithJson(w, 200, &returnedUser )
}
func (h Handlers) handlerRefresh(w http.ResponseWriter,r *http.Request){
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil{
		respondWithError(w, 401, "Invalid access token")
		return
	}
	if  tokenString == ""{
		respondWithError(w, 401, "Token string not found in headers.")
		return
	}
	refreshToken, err := h.apiConfig.db.GetRefreshToken(r.Context(), tokenString)
	if err != nil {
		respondWithError(w, 401, "Given Token not found")
		return
	}
	if !refreshToken.RevokedAt.Time.IsZero() || time.Now().After(refreshToken.ExpiresAt){
 		respondWithError(w, 401, "Given Token is expired")
		return
	}
	accessToken, err := auth.MakeJWT(refreshToken.UserID, h.apiConfig.tokenSecret, time.Duration(time.Hour * 1))
	if err != nil {
		fmt.Println("Error: Trying to generate access token for user.")
		fmt.Println(err)
		respondWithError(w, 500, "500 internal server error")
		return
	}
	returnJson := struct{
		Token string `json:"token"`
	}{Token: accessToken}
	respondWithJson(w, 200, &returnJson)
}
func (h Handlers) handlerRevoke(w http.ResponseWriter,r *http.Request){
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil{
		respondWithError(w, 401, "Invalid refresh token")
		return
	}
	if  tokenString == ""{
		respondWithError(w, 401, "Token string not found in headers.")
		return
	}
	refreshToken, err := h.apiConfig.db.GetRefreshToken(r.Context(), tokenString)
	if err != nil {
		respondWithError(w, 401, "Given Token not found")
		return
	}
	if !refreshToken.RevokedAt.Time.IsZero() {
		respondWithError(w, 401, "Given Token is expired")
		return
	}
	err = h.apiConfig.db.ExpireToken(r.Context(),refreshToken.Token)
	if err != nil {
		fmt.Println("Error: Trying to mark given token expire.")
		fmt.Println(err)
		respondWithError(w, 500, "Internal server error")
		return
	}
	respondWithJson(w, 204, &struct{}{})
}

func (h Handlers) handlerUpdateUser(w http.ResponseWriter,r *http.Request){
	reqBody:= struct{
		Email string `json:"email"`
		Password string  `json:"password"`
	}{}
	tokenString , err := auth.GetBearerToken(r.Header)
	userId, err := auth.ValidateJWT(tokenString, h.apiConfig.tokenSecret)
	if err != nil {
		respondWithError(w, 401, "Invalid token.")
		return
	}
	defer r.Body.Close()
	err = json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil{
		respondWithError(w, 400, "Invalid JSON.")
		return
	}
	if  reqBody.Email == "" || reqBody.Password == ""{
		respondWithError(w, 401, "Invalid request body. Must contain \"email\" \"password\".")
		return 
	}
	hashedPassword , err := auth.HashPassword(reqBody.Password)
	if err != nil {
		fmt.Println("Error: Trying to hash password")
		fmt.Println(err)
		respondWithError(w, 500, "Internal server error")
		return
	}
	updateUserParams := database.UpdateUserParams{
		ID: userId,
		Email: reqBody.Email,
		HashedPassword: hashedPassword,
	}
	updatedUser, err := h.apiConfig.db.UpdateUser(r.Context(), updateUserParams)
	if err != nil {
		fmt.Println("Error: Trying to update user data in database.")
		fmt.Println(err)
		respondWithError(w, 401, "Invalid access token")
		return
	}
	user := struct{
		ID        uuid.UUID `json:"id"`
		Email     string `json:"email"`
		IsChirpyRed bool `json:"is_chirpy_red"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`

	}{
		ID: updateUserParams.ID,
		Email: updateUserParams.Email,
		IsChirpyRed: updatedUser.IsChirpyRed,
		CreatedAt: updatedUser.CreatedAt,
		UpdatedAt: updatedUser.UpdatedAt,

	}
	respondWithJson(w, 200, &user)
}

func (h Handlers) handlerGetChirps(w http.ResponseWriter,r *http.Request){
	sortString := r.URL.Query().Get("sort")
	authorIdString := r.URL.Query().Get("author_id")
	var chirps []database.Chirp
	var err error
	authorId, err := uuid.Parse(authorIdString)
	if err == nil && authorId != uuid.Nil{
		chirps, err = h.apiConfig.db.GetChirpsForAuthor(r.Context(), authorId)
	}else {
		chirps, err = h.apiConfig.db.GetChirps(r.Context())
	}
	if err != nil {
		fmt.Println("Error: Trying to get chirps from database.")
		fmt.Println(err)
		respondWithError(w, 500, "500 Internal Server Error")
		return 
	}
	mappedChirps := make([]Chirp, len(chirps))
	for i,  chirp := range chirps{
		mappedChirps[i] = Chirp(chirp)
	}
	if sortString == "desc"{
		sort.Slice(mappedChirps, func(i, j int) bool { return mappedChirps[j].CreatedAt.Before( mappedChirps[i].CreatedAt)})
	}
	respondWithJson(w, 200, &mappedChirps)
}

func (h Handlers) handlerGetOneChirp(w http.ResponseWriter,r *http.Request){
	chirpIdString := r.PathValue("chirpID")
	if chirpIdString == ""{
		respondWithError(w, 400, "Request must include chirpID")
		return
	}
	chirpId, err:= uuid.Parse(chirpIdString)
	if err != nil {
		respondWithError(w, 400, "Invalid chirpID")
		return
	}
	chirp, err := h.apiConfig.db.GetChirp(r.Context(), chirpId)
	if err != nil {
		if err == sql.ErrNoRows{
			respondWithError(w, 404, "No data found for given chirpID")
			return 
		}
		fmt.Println("Error: Trying to get chirp for given chirpID from db.")
		fmt.Printf("ChirpId: %v\n", chirpId)
		fmt.Println(err)
		respondWithError(w, 500,"500 Internal Server Error ")
		return
	}
	mappedChirp := Chirp(chirp)
	respondWithJson(w, 200, &mappedChirp)
}

func (h Handlers) handlerDeleteChirp(w http.ResponseWriter,r *http.Request){
	tokenString , err := auth.GetBearerToken(r.Header)
	userId, err := auth.ValidateJWT(tokenString, h.apiConfig.tokenSecret)
	if err != nil {
		respondWithError(w, 401, "Invalid access token.")
		return
	}
	chirpIdString := r.PathValue("chirpID")
	if chirpIdString == "" {
		respondWithError(w, 404, "chirpID not found in request")
		return
	}
	chirpId, err := uuid.Parse(chirpIdString)
	if err != nil {
		respondWithError(w, 401, "Invalid chirpID.")
		return
	}
	chirpToBeDeleted, err:= h.apiConfig.db.GetChirp(context.Background(), chirpId)
	if err != nil {
		if err == sql.ErrNoRows{
			respondWithError(w, 403, "Given ChirpId Does not exist for loggen in user.")
			return
		}
		fmt.Println(err)
		respondWithError(w, 500, "Internal server Error")
		return 

	}

	if userId != chirpToBeDeleted.UserID{
		respondWithError(w, 403, "Given ChirpId Does not exist for loggen in user.")
		return 
	}
	deleteChirpParams :=  database.DeleteChirpParams{
		ID: chirpId,
		UserID: userId,
	}
	err = h.apiConfig.db.DeleteChirp(r.Context(), deleteChirpParams)
	if err != nil {
		fmt.Println(r.URL.Path)
		fmt.Println("Error: Trying to update chirp in database.")
		fmt.Println(err)
		respondWithError(w, 500, "Internal server Error")
		return
	}
	respondWithJson(w, 204, &struct{}{})
}

func (h Handlers) handlerUpgradeUserToRed(w http.ResponseWriter,r *http.Request){
	reqBody := struct{
		Event string `json:"event"`
		Data struct {
			UserId string `json:"user_id"`
		} `json:"data"`
	}{}
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil  || apiKey != h.apiConfig.polkaKey{
		respondWithError(w, 401, "Invalid or unknown apiKey in auth headers")
	}
	defer r.Body.Close()
	err = json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		respondWithError(w, 400, "Invalid json")
		return
	}
	if reqBody.Event != "user.upgraded"{
		respondWithJson(w, 204, &struct{}{})
		return
	}
	userId ,err := uuid.Parse(reqBody.Data.UserId)
	if err != nil {
		respondWithError(w, 400, "Invalid user_id")
		return
	}
	err = h.apiConfig.db.UpgradeUserToRed(r.Context(), userId)
	if err != nil {
		if err == sql.ErrNoRows{
			respondWithError(w, 404, "User not found")
			return
		}
		respondWithError(w, 500, "Internal Server Error")
		return
	}
	respondWithJson(w, 204, &struct{}{})
}
