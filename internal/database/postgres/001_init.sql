create table categories (category_id serial, category_name text unique); 
create table products (product_id serial, product_name text unique, category text, stock int);

---- create above / drop below ----

drop table categories, products;