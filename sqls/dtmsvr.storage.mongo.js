use dtm
db.trans.drop()

db.trans.createIndex({gid:1},{unique: true })
db.trans.createIndex({owner:1})
db.trans.createIndex({status:1, next_cron_time:1}, {sparse:true})
db.trans.createIndex({ gid: 1, branch_trans.branch_id: 1, branch_trans.op: 1}, { unique: true })