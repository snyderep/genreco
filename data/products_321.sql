SELECT account_id,
       pid,
       name,
       product_url,
       url,
       unit_price
INTO OUTFILE '/tmp/productts.txt'
FROM retailer_product
WHERE account_id = 321
ORDER BY account_id, pid;
