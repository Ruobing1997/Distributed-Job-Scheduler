docker tag morefun-supernova-api:4.0 mirrors.tencent.com/supernova/supernova-repo:api-server
docker tag morefun-supernova-manager:4.0 mirrors.tencent.com/supernova/supernova-repo:manager
docker tag morefun-supernova-worker:4.0 mirrors.tencent.com/supernova/supernova-repo:worker
docker push mirrors.tencent.com/supernova/supernova-repo:worker
docker push mirrors.tencent.com/supernova/supernova-repo:manager
docker push mirrors.tencent.com/supernova/supernova-repo:api-server