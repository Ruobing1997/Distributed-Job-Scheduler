docker tag morefun-supernova-api:3.0 mirrors.tencent.com/supernova/supernova-repo:api
docker tag morefun-supernova-manager:3.0 mirrors.tencent.com/supernova/supernova-repo:manager
docker tag morefun-supernova-worker:3.0 mirrors.tencent.com/supernova/supernova-repo:worker
docker push mirrors.tencent.com/supernova/supernova-repo:worker
docker push mirrors.tencent.com/supernova/supernova-repo:manager
docker push mirrors.tencent.com/supernova/supernova-repo:api