CREATE SCHEMA if not exists dtm_busi
/* SQLINES DEMO *** RACTER SET utf8mb4 */
;
drop table if exists dtm_busi.user_account;
-- SQLINES LICENSE FOR EVALUATION USE ONLY
create sequence if not exists dtm_busi.user_account_seq;
create table if not exists dtm_busi.user_account(
  id int PRIMARY KEY DEFAULT NEXTVAL ('dtm_busi.user_account_seq'),
  user_id int UNIQUE,
  balance DECIMAL(10, 2) not null default '0',
  trading_balance DECIMAL(10, 2) not null default '0',
  create_time timestamp(0) with time zone DEFAULT now(),
  update_time timestamp(0) with time zone DEFAULT now()
);
-- SQLINES LICENSE FOR EVALUATION USE ONLY
create index if not exists create_idx on dtm_busi.user_account(create_time);
-- SQLINES LICENSE FOR EVALUATION USE ONLY
create index if not exists update_idx on dtm_busi.user_account(update_time);
TRUNCATE dtm_busi.user_account;
insert into dtm_busi.user_account (user_id, balance)
values (1, 10000),
  (2, 10000);