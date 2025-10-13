insert into attachments (inspection_id, type, file_id)
values (:inspection_id, :type, :file_id)
returning id, inspection_id, type, file_id, created_at;
