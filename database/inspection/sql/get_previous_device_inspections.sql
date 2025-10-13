select id, device_id, inspection_id, value, consumption, created_at
from inspected_devices
where device_id = $1
  and inspection_id != $2
order by created_at desc;
