ADD JAR s3://dev-emr.monetate.net/lib/hive-json-serde-0.3.jar;
SET hive.optimize.s3.query=true;

SET mapred.job.name=EPS-HACK-USER-PRODUCTS-PURCHASED;

SELECT p.account_id AS account_id, p.product_id AS pid, COUNT(pp.quantity) / COUNT(p.views) AS conv
FROM product p LEFT OUTER JOIN purchase_product pp ON (
    p.product_id = pp.product_id AND
    p.account_id = pp.account_id AND
    p.dt = pp.dt)
WHERE p.account_id = 321 
AND   p.dt >= 20121209 
AND   p.dt <= 20121216
GROUP BY p.account_id, p.product_id
ORDER BY account_id, pid; 
