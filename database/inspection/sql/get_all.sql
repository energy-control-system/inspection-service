select id,
       task_id,
       status,
       type,
       resolution,
       limit_reason,
       method,
       method_by,
       reason_type,
       reason_description,
       is_restriction_checked,
       is_violation_detected,
       is_expense_available,
       violation_description,
       is_unauthorized_consumers,
       unauthorized_description,
       unauthorized_explanation,
       inspect_at,
       energy_action_at,
       created_at,
       updated_at
from inspections
order by
    case when $1 = 'asc' then inspect_at end asc nulls last,
    case when $1 = 'desc' then inspect_at end desc nulls last,
    id
limit $2 offset $3;
