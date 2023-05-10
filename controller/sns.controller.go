package controller

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/04Akaps/Jenkins_docker_go.git/crypto"
	connection "github.com/04Akaps/Jenkins_docker_go.git/mysql"
	sqlc "github.com/04Akaps/Jenkins_docker_go.git/mysql/sqlc"
	"github.com/04Akaps/Jenkins_docker_go.git/utils"
	"github.com/gorilla/mux"
)

type SnsController struct {
	Ctx         context.Context
	MySQLClient *sqlc.Queries
	EthClient   crypto.CryptoClientImpl
}

type SnsImpl interface {
	GetSnsByID(http.ResponseWriter, *http.Request)
	GetAllSnsByEoaAddress(http.ResponseWriter, *http.Request)
	MakeSns(http.ResponseWriter, *http.Request)
}

func NewSnsController() SnsImpl {
	context := context.Background()
	endPoint := "https://mainnet.infura.io/v3/299623e5cf3442c8bb2dbe870d8f7d88"
	// 어차피 개인 프로젝트이기 떄문에 Fix

	client := crypto.NewEthClient(context, endPoint)

	return &SnsController{Ctx: context, MySQLClient: connection.NewMySQLClient("sns"), EthClient: client}
}

func (sc *SnsController) GetSnsByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	log.Println("GetSnsByID", id)

	// id에 해당하는 sns를 념거주자
	numId, err := strconv.Atoi(id)
	if err != nil {
		// 변환 실패
		log.Println("Atoi 변환 실패")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if numId < 1 {
		log.Println("존재하지 않는 Key")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result, err := sc.MySQLClient.GetSnsPost(sc.Ctx, int64(numId))
	if err != nil {
		log.Println("Get Query Failed --> ", err)
		w.WriteHeader(http.StatusConflict)
		return
	}

	log.Println(result)

	w.WriteHeader(http.StatusOK)
}

func (sc *SnsController) GetAllSnsByEoaAddress(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["eoaAddress"]

	if !sc.EthClient.IsEoaAddress(address) || sc.EthClient.IsContractAddress(sc.Ctx, address) {
		log.Println("16진수가 아니고 40글자도 아닌경우 & Contract 주소 인 경우")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 이제 모든 데이터를 가져오면 된다.

	result, err := sc.MySQLClient.GetSnsPostAll(sc.Ctx, address)
	if err != nil {
		log.Println("Get All Query Failed --> ", err)
		w.WriteHeader(http.StatusConflict)
		return
	}

	log.Println(result)

	w.WriteHeader(http.StatusOK)
}

func (sc *SnsController) MakeSns(w http.ResponseWriter, r *http.Request) {
	log.Println("MakeSns")

	var req sqlc.CreateNewSnsPostParams
	decoder := utils.BodyDecoder(w, r)

	if err := decoder.Decode(&req); err != nil {
		log.Println("디코딩 실패")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 주소 검증
	if !sc.EthClient.IsEoaAddress(req.PostOwnerAccount) || sc.EthClient.IsContractAddress(sc.Ctx, req.PostOwnerAccount) {
		log.Println("16진수가 아니고 40글자도 아닌경우 & Contract 주소 인 경우")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(req.Text) == 0 {
		// 글 내용이 없는 경우
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(req.Title) == 0 {
		// 제목이 없는 경우
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 이미지를 파일로 받아서 처리를 하는것이 일반적이고 효율적으로 알고 있는데,
	// 해당 부분을 하는 방법을 몰라서... 일단 base64로 저장

	_, err := sc.MySQLClient.CreateNewSnsPost(sc.Ctx, req)
	if err != nil {
		// DB Insert 실패
		log.Println("Insert Query Failed --> ", err)
		w.WriteHeader(http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusOK)
}
