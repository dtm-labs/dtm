CREATE SCHEMA if not exists dtm_busi /* SQLINES DEMO *** RACTER SET utf8mb4 */;
create SCHEMA  if not exists dtm_barrier /* SQLINES DEMO *** RACTER SET utf8mb4 */;

drop table if exists dtm_busi.user_account;
-- SQLINES LICENSE FOR EVALUATION USE ONLY
create sequence if not exists dtm_busi.user_account_seq;

create table if not exists dtm_busi.user_account(
  id int PRIMARY KEY DEFAULT NEXTVAL ('dtm_busi.user_account_seq'),
  user_id int UNIQUE ,
  balance DECIMAL(10, 2) not null default '0',
  create_time timestamp(0) DEFAULT now(),
  update_time timestamp(0) DEFAULT now()
);
-- SQLINES LICENSE FOR EVALUATION USE ONLY
create index if not exists create_idx on dtm_busi.user_account(create_time);
-- SQLINES LICENSE FOR EVALUATION USE ONLY
create index if not exists update_idx on dtm_busi.user_account(update_time);

TRUNCATE dtm_busi.user_account
insert into dtm_busi.user_account (user_id, balance) values (1, 10000), (2, 10000);

drop table if exists dtm_busi.user_account_trading;
-- SQLINES LICENSE FOR EVALUATION USE ONLY
create sequence if not exists dtm_busi.user_account_trading_seq;

create table if not exists dtm_busi.user_account_trading( -- SQLINES DEMO *** �冻结的金额
  id int PRIMARY KEY DEFAULT NEXTVAL ('dtm_busi.user_account_trading_seq'),
  user_id int UNIQUE ,
  trading_balance DECIMAL(10, 2) not null default '0',
  create_time timestamp(0) DEFAULT now(),
  update_time timestamp(0) DEFAULT now()
);
-- SQLINES LICENSE FOR EVALUATION USE ONLY
create index if not exists create_idx on dtm_busi.user_account_trading(create_time);
-- SQLINES LICENSE FOR EVALUATION USE ONLY
create index if not exists update_idx on dtm_busi.user_account_trading(update_time);

TRUNCATE dtm_busi.user_account_trading;
insert into dtm_busi.user_account_trading (user_id, trading_balance) values (1, 0), (2, 0);


drop table if exists dtm_busi.barrier;
-- SQLINES LICENSE FOR EVALUATION USE ONLY
create sequence if not exists dtm_busi.barrier_seq;

create table if not exists dtm_busi.barrier(
  id int PRIMARY KEY DEFAULT NEXTVAL ('dtm_busi.barrier_seq'),
  trans_type varchar(45) default '' ,
  gid varchar(128) default'',
  branch_id varchar(128) default '',
  branch_type varchar(45) default '',
  reason varchar(45) default '' ,
  result varchar(2047) default null ,
  create_time timestamp(0) DEFAULT now(),
  update_time timestamp(0) DEFAULT now(),
  UNIQUE (gid, branch_id, branch_type)
);
-- SQLINES LICENSE FOR EVALUATION USE ONLY
create index if not exists create_idx on dtm_busi.barrier(create_time);
-- SQLINES LICENSE FOR EVALUATION USE ONLY
create index if not exists update_idx on dtm_busi.barrier(update_time);
