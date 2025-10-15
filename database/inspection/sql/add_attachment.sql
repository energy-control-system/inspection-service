insert into attachments (inspection_id, type, file_id)
values ($1, $2, $3)
returning id, inspection_id, type, file_id, created_at;
