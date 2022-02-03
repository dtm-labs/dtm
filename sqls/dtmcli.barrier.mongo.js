use dtm_barrier
db.barrier.drop()
db.barrier.createIndex({gid:1, branch_id:1, op: 1, barrier_id: 1}, {unique: true})
//db.barrier.insert({gid:"123", branch_id:"01", op:"action", barrier_id:"01", reason:"action"});
