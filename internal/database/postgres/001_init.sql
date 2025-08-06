create table categories (category_id serial primary key, category_name text unique); 
create table products (product_id serial primary key, product_name text unique, category_id int references categories(category_id), stock int);

---- create above / drop below ----

drop table categories, products;