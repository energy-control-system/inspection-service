select id, inspection_id, type, file_id, created_at
from attachments
where inspection_id in (?)
order by id;
