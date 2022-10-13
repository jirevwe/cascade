# Set version
tag=$1
: > ./internal/pkg/version/VERSION && \
echo $tag > ./internal/pkg/version/VERSION

# Commit version number & push
git add ./internal/pkg/version/VERSION
git commit -m "Bump version to $tag"
git push origin

# Tag & Push.
git tag $tag
git push origin $tag
