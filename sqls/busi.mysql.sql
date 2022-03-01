CREATE DATABASE if not exists dtm_busi
/*!40100 DEFAULT CHARACTER SET utf8mb4 */
;
drop table if exists dtm_busi.user_account;
create table if not exists dtm_busi.user_account(
  id int(11) PRIMARY KEY AUTO_INCREMENT,
  user_id int(11) UNIQUE,
  balance DECIMAL(10, 2) not null default '0',
  trading_balance DECIMAL(10, 2) not null default '0',
  create_time datetime DEFAULT now(),
  update_time datetime DEFAULT now(),
  key(create_time),
  key(update_time)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
insert into dtm_busi.user_account (user_id, balance)
values (1, 10000),
  (2, 10000) on DUPLICATE KEY
UPDATE balance =
values (balance);