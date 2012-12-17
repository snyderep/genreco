ADD JAR s3://dev-emr.monetate.net/lib/hive-json-serde-0.3.jar;
SET hive.optimize.s3.query=true;

SET mapred.job.name=EPS-HACK-USER-PRODUCTS-VIEWED;

SELECT account_id, monetate_id, product_id, SUM(views)
FROM product
WHERE account_id = 321 
AND   dt >= 20121209 
AND   dt <= 20121216
GROUP BY account_id, monetate_id, product_id
ORDER BY account_id, monetate_id, product_id; 
