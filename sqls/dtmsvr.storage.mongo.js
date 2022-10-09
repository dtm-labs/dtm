use dtm
db.trans.drop()

db.trans.createIndex({gid:1},{unique: true })
db.trans.createIndex({owner:1})
db.trans.createIndex({status:1, next_cron_time:1}, {sparse:true})
db.trans.createIndex({ gid: 1, branch_trans.branch_id: 1, branch_trans.op: 1}, { unique: true })
db.trans.insert({"gid":"oijda","create_time":"2006-01-02T15:04:05.999Z","update_time":"2006-01-02T15:04:05.999Z","trans_type":"type1","query_prepared":"query_prepared1","protocol":"protocol1","finish_time":"2006-01-02T15:04:05.999Z","rollback_time":"2006-01-02T15:04:05.999Z","options":"ops","custom_data":"cd","next_cron_interval":2,"next_cron_time":"2006-01-02T15:04:05.999Z","owner":"owner1","ext_data":"ext_data1","branch_trans":[{"branch_id" : "branchIdoi9","branch_create_time":"2006-01-02T15:04:05.999Z","branch_update_time":"2006-01-02T15:04:05.999Z","branch_finish_time":"2006-01-02T15:04:05.999Z","branch_rollback_time":"2006-01-02T15:04:05.999Z",},{"branch_id" : "branchIdlo1",}]})