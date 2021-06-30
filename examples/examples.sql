CREATE DATABASE if not exists `dtm_busi` /*!40100 DEFAULT CHARACTER SET utf8mb4 */;
use dtm_busi;
drop table if exists user_account;
create table if not exists user_account(
  id int(11) PRIMARY KEY AUTO_INCREMENT,
  user_id int(11) UNIQUE ,
  balance DECIMAL(10, 2) not null default '0',
  create_time datetime DEFAULT now(),
  update_time datetime DEFAULT now(),
  key(create_time),
  key(update_time)
);

insert into user_account (user_id, balance) values (1, 10000), (2, 10000) on DUPLICATE KEY UPDATE balance=values (balance);

drop table if exists user_account_trading;
create table if not exists user_account_trading( -- 表示交易中被冻结的金额
  id int(11) PRIMARY KEY AUTO_INCREMENT,
  user_id int(11) UNIQUE ,
  trading_balance DECIMAL(10, 2) not null default '0',
  create_time datetime DEFAULT now(),
  update_time datetime DEFAULT now(),
  key(create_time),
  key(update_time)
);

insert into user_account_trading (user_id, trading_balance) values (1, 0), (2, 0) on DUPLICATE KEY UPDATE trading_balance=values (trading_balance);