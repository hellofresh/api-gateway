package janus

import (
	log "github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2"
)

// APILoader will load an Api definition from a storage system. It has two methods LoadDefinitionsFromMongo()
// and LoadDefinitions(), each will pull api specifications from different locations.
type APILoader struct{}

func (a *APILoader) LoadDefinitions(dir string) {

}

func (a *APILoader) LoadDefinitionsFromDatastore(dbSession *mgo.Session) []*APISpec {
	repo, err := NewMongoAppRepository(dbSession.DB(""))
	if err != nil {
		log.Panic(err)
	}

	definitions, err := repo.FindAll()
	if err != nil {
		log.Panic(err)
	}

	var APISpecs = []*APISpec{}
	for _, definition := range definitions {
		newAppSpec := APISpec{}
		newAppSpec.APIDefinition = definition
		APISpecs = append(APISpecs, &newAppSpec)
	}

	return APISpecs
}

func (a *APILoader) LoadOauthServersFromDatastore(dbSession *mgo.Session) []*OAuthSpec {
	repo, err := NewMongoOAuthRepository(dbSession.DB(""))
	if err != nil {
		log.Panic(err)
	}

	oauthServers, err := repo.FindAll()
	if err != nil {
		log.Panic(err)
	}

	var oAuthSpecs []*OAuthSpec
	for _, oauth := range oauthServers {
		var spec OAuthSpec
		spec.OAuth = oauth
		oAuthSpecs = append(oAuthSpecs, &spec)
	}

	return oAuthSpecs
}