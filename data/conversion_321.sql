SELECT pid, other_pid
INTO OUTFILE '/tmp/product_conversion.txt'
FROM product_relationship pr
WHERE account_id = 321
AND relationship_id = 2
ORDER BY account_id, pid, other_pid;
