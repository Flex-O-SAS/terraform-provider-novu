Ceci est un WIP, qui sert aussi à apprendre le framework Terraform Plugin Framework et le langage Go.

# TODO : 
- Ajouter les tests
- Ajouter les exemples
- Passer sur la CI
- Gérer le déploiement sur le registry
- Régler les todo dans le code
- Tester la génération des docs


## Questions à régler : 
- Le SDK a régulièrement des breaking changes -> est-ce qu'on fixe la version ?

- Est-ce utile d'avoir des champs inutilisés dans le provider s'ils sont dispo dans l'API ?
    - owner_id pour les api keys
    - key sha pour les api keys
    - organization_id pour les environments

- Est-ce qu'on réutilise les structures entre data sources ou bien on redéclare tout à chaque fois ?
ex1 : type EnvironmentWithKeysDataSourceModel struct {
	ApiKeys []ApiKeyDataSourceModel `tfsdk:"api_keys"`
	EnvironmentDataSourceModel
}

ex 2 : type EnvironmentsWithKeysDataSourceModel struct {
	Items []EnvironmentWithKeysDataSourceModel `tfsdk:"items"`
}

- est-ce qu'on fait en sorte d'avoir le même datamodel exact entre resource et data source ?

- Y a-t-il une raison à avoir à la fois une source et une ressource pour un même objet?

- On déclare chaque struct dans le fichier le plus pertinent, ou à un endroit central ?

- une resource = un fichier, ou on peut split ?

- novu providers data source : mon implem est ultra crade, est-ce qu'on fait plutôt une ressource par provider qu'on veut gérer 
plutôt qu'un truc générique immonde ?

- pour les environnements ressources, est-ce qu'il ne faudrait pas filtrer sur l'ID si il est set plutôt que sur le nom ?

- Au read, on update tous les fields pour être sûr ?

- à l'update, on update que les fields qui ont changé ?

- à l'update, on refait un read pour être sûr que les fields sont corrects ?



## NB : 
This repo uses a fork of the repo github.com/novuhq/novu-go to add direct API call functionality, when the SDK resources are broken.
It is set up with a go mod replace that points to said fork.
To remove this replace, use `go mod edit -dropreplace=github.com/novuhq/novu-go`

## Novu SDK Version
Be careful before upgrading the novu SDK :
1) It has breaking changes even in minor versions
2) It regularly introduces critical bugs in new versions


Links to the novu go docs are generated based on the version of the novugo package. If the package version does not equate a github tag, they will be invalid.