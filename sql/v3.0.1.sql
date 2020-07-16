set names utf8;
use n9e_uic;
ALTER TABLE team ADD COLUMN `nid` int unsigned NOT NULL COMMENT '关联服务树id';

use n9e_mon;
ALTER TABLE event_cur ADD COLUMN `alert_users` VARCHAR(1000) NOT NULL COMMENT '已报警用户';

use n9e_uic;
alter table team drop index ident;
CREATE UNIQUE INDEX u_ident_nid ON team (ident,nid);

