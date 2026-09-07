find . -type f -not -path './**/vendor/*' -name 'go.mod' -print0 | 
    while IFS= read -r -d '' line; do 
        dirPath=${line%/*}/
        echo ${dirPath}
##        (cd ${dirPath} && ls && go mod vendor)
        (cd ${dirPath} && ls && rm -r go.mod && rm -r go.sum && rm -r -f vendor && go mod init && go mod tidy)
    done