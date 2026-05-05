select id, device_id, inspection_id, value, consumption, created_at
from inspected_devices
where inspection_id = $1
order by id;
