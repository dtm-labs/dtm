use dtm_busi
db.user_account.drop()
db.user_account.createIndex({ user_id: NumberLong(1) }, { unique: true })

db.user_account.insert({ user_id: NumberLong(1), balance: 10000 })
db.user_account.insert({ user_id: NumberLong(2), balance: 10000 })
