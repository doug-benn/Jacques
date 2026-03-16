CREATE TYPE user_status AS ENUM ('active', 'inactive', 'suspended', 'pending');

CREATE TYPE account_type AS ENUM ('personal', 'business', 'enterprise', 'trial');

CREATE TYPE order_status AS ENUM ('pending', 'processing', 'shipped', 'delivered', 'cancelled', 'refunded');

CREATE TYPE payment_method AS ENUM ('credit_card', 'debit_card', 'paypal', 'bank_transfer', 'crypto');

CREATE TYPE payment_status AS ENUM ('pending', 'completed', 'failed', 'refunded', 'disputed');

CREATE TYPE notification_type AS ENUM ('email', 'sms', 'push', 'in_app');

CREATE TYPE log_level AS ENUM ('debug', 'info', 'warning', 'error', 'critical');

CREATE SEQUENCE global_id_seq START WITH 1 INCREMENT BY 1 NO MAXVALUE CACHE 1;

CREATE TABLE organizations (
    id bigint NOT NULL PRIMARY KEY DEFAULT nextval('global_id_seq'::regclass),
    name text NOT NULL,
    slug text NOT NULL UNIQUE,
    account_type account_type NOT NULL DEFAULT 'trial',
    max_users integer NOT NULL DEFAULT 10,
    created_at timestamp without time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE users (
    id bigint NOT NULL PRIMARY KEY DEFAULT nextval('global_id_seq'::regclass),
    organization_id bigint  REFERENCES organizations(id),
    email text NOT NULL UNIQUE,
    password_hash text NOT NULL,
    first_name text NOT NULL,
    last_name text NOT NULL,
    phone text,
    avatar_url text,
    status user_status NOT NULL DEFAULT 'pending',
    email_verified_at timestamp without time zone,
    last_login_at timestamp without time zone,
    failed_login_attempts integer NOT NULL DEFAULT 0,
    locked_until timestamp without time zone,
    created_at timestamp without time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp without time zone NOT NULL DEFAULT NOW(),
    UNIQUE (email, organization_id)
);

CREATE TABLE user_profiles (
    user_id bigint NOT NULL PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    bio text,
    date_of_birth date,
    gender text,
    address_line1 text,
    address_line2 text,
    city text,
    state text,
    postal_code text,
    country text NOT NULL DEFAULT 'US',
    timezone text NOT NULL DEFAULT 'UTC',
    locale text NOT NULL DEFAULT 'en-US',
    preferences jsonb NOT NULL DEFAULT '{}'
);

CREATE TABLE sessions (
    id bigint NOT NULL PRIMARY KEY,
    user_id bigint  REFERENCES users(id) ON DELETE CASCADE,
    token text NOT NULL UNIQUE,
    ip_address inet,
    user_agent text,
    expires_at timestamp without time zone NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE password_resets (
    id bigint NOT NULL PRIMARY KEY,
    user_id bigint  REFERENCES users(id) ON DELETE CASCADE,
    token text NOT NULL UNIQUE,
    used_at timestamp without time zone,
    expires_at timestamp without time zone NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE categories (
    id bigint NOT NULL PRIMARY KEY DEFAULT nextval('global_id_seq'::regclass),
    parent_id bigint REFERENCES categories(id),
    name text NOT NULL,
    slug text NOT NULL UNIQUE,
    description text,
    display_order integer NOT NULL DEFAULT 0,
    is_active boolean NOT NULL DEFAULT true,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE products (
    id bigint NOT NULL PRIMARY KEY DEFAULT nextval('global_id_seq'::regclass),
    category_id bigint REFERENCES categories(id),
    sku text NOT NULL UNIQUE,
    name text NOT NULL,
    slug text NOT NULL UNIQUE,
    description text,
    price numeric(10,2) NOT NULL,
    cost numeric(10,2),
    weight numeric(10,2),
    is_active boolean NOT NULL DEFAULT true,
    is_featured boolean NOT NULL DEFAULT false,
    meta_title text,
    meta_description text,
    created_at timestamp without time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp without time zone NOT NULL DEFAULT NOW(),
    CHECK (price >= 0)
);

CREATE TABLE product_variants (
    id bigint NOT NULL PRIMARY KEY,
    product_id bigint  REFERENCES products(id) ON DELETE CASCADE,
    sku text NOT NULL UNIQUE,
    name text NOT NULL,
    price numeric(10,2),
    stock_quantity integer NOT NULL DEFAULT 0,
    low_stock_threshold integer NOT NULL DEFAULT 10,
    is_active boolean NOT NULL DEFAULT true,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE warehouses (
    id bigint NOT NULL PRIMARY KEY,
    name text NOT NULL,
    code text NOT NULL UNIQUE,
    address_line1 text NOT NULL,
    address_line2 text,
    city text NOT NULL,
    state text NOT NULL,
    postal_code text NOT NULL,
    country text NOT NULL DEFAULT 'US',
    is_primary boolean NOT NULL DEFAULT false,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE inventory (
    id bigint NOT NULL PRIMARY KEY,
    product_id bigint  REFERENCES products(id) ON DELETE CASCADE,
    variant_id bigint REFERENCES product_variants(id) ON DELETE CASCADE,
    warehouse_id bigint  REFERENCES warehouses(id),
    quantity integer NOT NULL DEFAULT 0,
    reserved_quantity integer NOT NULL DEFAULT 0,
    updated_at timestamp without time zone NOT NULL DEFAULT NOW(),
    CHECK (quantity >= 0)
);

CREATE TABLE suppliers (
    id bigint NOT NULL PRIMARY KEY,
    name text NOT NULL,
    code text NOT NULL UNIQUE,
    contact_name text,
    email text,
    phone text,
    address_line1 text,
    city text,
    state text,
    postal_code text,
    country text NOT NULL DEFAULT 'US',
    payment_terms text,
    lead_time_days integer,
    is_active boolean NOT NULL DEFAULT true,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE orders (
    id bigint NOT NULL PRIMARY KEY,
    user_id bigint  REFERENCES users(id),
    order_number text NOT NULL UNIQUE,
    status order_status NOT NULL DEFAULT 'pending',
    subtotal numeric(10,2) NOT NULL,
    tax_amount numeric(10,2) NOT NULL DEFAULT 0,
    shipping_amount numeric(10,2) NOT NULL DEFAULT 0,
    discount_amount numeric(10,2) NOT NULL DEFAULT 0,
    total_amount numeric(10,2) NOT NULL,
    currency text NOT NULL DEFAULT 'USD',
    shipping_address_line1 text,
    shipping_address_line2 text,
    shipping_city text,
    shipping_state text,
    shipping_postal_code text,
    shipping_country text,
    billing_address_line1 text,
    billing_address_line2 text,
    billing_city text,
    billing_state text,
    billing_postal_code text,
    billing_country text,
    notes text,
    placed_at timestamp without time zone NOT NULL DEFAULT NOW(),
    shipped_at timestamp without time zone,
    delivered_at timestamp without time zone,
    cancelled_at timestamp without time zone,
    created_at timestamp without time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp without time zone NOT NULL DEFAULT NOW(),
    CHECK (total_amount >= 0)
);

CREATE TABLE order_items (
    id bigint NOT NULL PRIMARY KEY,
    order_id bigint  REFERENCES orders(id) ON DELETE CASCADE,
    product_id bigint  REFERENCES products(id),
    variant_id bigint REFERENCES product_variants(id),
    sku text NOT NULL,
    name text NOT NULL,
    quantity integer NOT NULL,
    unit_price numeric(10,2) NOT NULL,
    discount_amount numeric(10,2) NOT NULL DEFAULT 0,
    tax_amount numeric(10,2) NOT NULL DEFAULT 0,
    total_amount numeric(10,2) NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT NOW(),
    CHECK (quantity > 0)
);

CREATE TABLE payments (
    id bigint NOT NULL PRIMARY KEY,
    order_id bigint  REFERENCES orders(id),
    transaction_id text UNIQUE,
    payment_method payment_method NOT NULL,
    status payment_status NOT NULL DEFAULT 'pending',
    amount numeric(10,2) NOT NULL,
    currency text NOT NULL DEFAULT 'USD',
    gateway_response text,
    failure_reason text,
    processed_at timestamp without time zone,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE departments (
    id bigint NOT NULL PRIMARY KEY,
    organization_id bigint  REFERENCES organizations(id),
    parent_id bigint REFERENCES departments(id),
    name text NOT NULL,
    code text NOT NULL,
    manager_id bigint,
    budget numeric(12,2),
    created_at timestamp without time zone NOT NULL DEFAULT NOW(),
    UNIQUE (organization_id, code)
);

CREATE TABLE employees (
    id bigint NOT NULL PRIMARY KEY,
    user_id bigint REFERENCES users(id),
    department_id bigint REFERENCES departments(id),
    employee_number text NOT NULL UNIQUE,
    job_title text NOT NULL,
    hire_date date NOT NULL,
    termination_date date,
    employment_type text NOT NULL DEFAULT 'full-time',
    hourly_rate numeric(10,2),
    is_active boolean NOT NULL DEFAULT true,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE salaries (
    id bigint NOT NULL PRIMARY KEY,
    employee_id bigint  REFERENCES employees(id),
    effective_date date NOT NULL,
    salary_amount numeric(12,2) NOT NULL,
    bonus_amount numeric(12,2) DEFAULT 0,
    currency text NOT NULL DEFAULT 'USD',
    pay_frequency text NOT NULL DEFAULT 'monthly',
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE posts (
    id bigint NOT NULL PRIMARY KEY,
    author_id bigint  REFERENCES users(id),
    title text NOT NULL,
    slug text NOT NULL UNIQUE,
    content text,
    excerpt text,
    featured_image_url text,
    status text NOT NULL DEFAULT 'draft',
    published_at timestamp without time zone,
    view_count integer NOT NULL DEFAULT 0,
    created_at timestamp without time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE comments (
    id bigint NOT NULL PRIMARY KEY,
    post_id bigint REFERENCES posts(id) ON DELETE CASCADE,
    parent_id bigint REFERENCES comments(id),
    author_id bigint  REFERENCES users(id),
    content text NOT NULL,
    is_approved boolean NOT NULL DEFAULT true,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE likes (
    id bigint NOT NULL PRIMARY KEY,
    user_id bigint  REFERENCES users(id),
    post_id bigint REFERENCES posts(id) ON DELETE CASCADE,
    comment_id bigint REFERENCES comments(id) ON DELETE CASCADE,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE follows (
    id bigint NOT NULL PRIMARY KEY,
    follower_id bigint  REFERENCES users(id),
    following_id bigint  REFERENCES users(id),
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE settings (
    id bigint NOT NULL PRIMARY KEY,
    organization_id bigint REFERENCES organizations(id),
    key text NOT NULL,
    value text,
    data_type text NOT NULL DEFAULT 'string',
    is_encrypted boolean NOT NULL DEFAULT false,
    description text,
    created_at timestamp without time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE notifications (
    id bigint NOT NULL PRIMARY KEY,
    user_id bigint  REFERENCES users(id),
    type notification_type NOT NULL,
    title text NOT NULL,
    message text NOT NULL,
    data jsonb,
    is_read boolean NOT NULL DEFAULT false,
    read_at timestamp without time zone,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE audit_logs (
    id bigint NOT NULL PRIMARY KEY,
    organization_id bigint REFERENCES organizations(id),
    user_id bigint REFERENCES users(id),
    action text NOT NULL,
    entity_type text NOT NULL,
    entity_id bigint,
    changes jsonb,
    ip_address inet,
    user_agent text,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE api_keys (
    id bigint NOT NULL PRIMARY KEY,
    organization_id bigint  REFERENCES organizations(id),
    name text NOT NULL,
    key_hash text NOT NULL UNIQUE,
    prefix text NOT NULL,
    last_used_at timestamp without time zone,
    expires_at timestamp without time zone,
    is_active boolean NOT NULL DEFAULT true,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE VIEW active_users AS
SELECT u.id, u.email, u.first_name, u.last_name, u.status, o.name as organization_name
FROM users u
JOIN organizations o ON u.organization_id = o.id
WHERE u.status = 'active';

CREATE VIEW order_summary AS
SELECT o.id, o.order_number, o.total_amount, o.status, o.created_at,
       u.email as customer_email, u.first_name || ' ' || u.last_name as customer_name
FROM orders o
JOIN users u ON o.user_id = u.id;

CREATE VIEW product_inventory_summary AS
SELECT p.id, p.name, p.sku, p.price,
       COALESCE(SUM(i.quantity), 0) as total_quantity,
       COUNT(DISTINCT w.id) as warehouse_count
FROM products p
LEFT JOIN inventory i ON p.id = i.product_id
LEFT JOIN warehouses w ON i.warehouse_id = w.id
GROUP BY p.id, p.name, p.sku, p.price;

CREATE MATERIALIZED VIEW order_stats AS
SELECT status, COUNT(*) as count, COALESCE(SUM(total_amount), 0) as total_amount
FROM orders
GROUP BY status;

CREATE INDEX idx_order_stats_status ON order_stats(status);

CREATE MATERIALIZED VIEW user_order_summary AS
SELECT u.id as user_id, u.email, COUNT(o.id) as order_count
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
GROUP BY u.id, u.email;

CREATE INDEX idx_user_order_user_id ON user_order_summary(user_id);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION set_current_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.created_at = NOW();
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION get_order_total(order_id bigint)
RETURNS numeric AS $$
DECLARE
    total numeric(10,2);
BEGIN
    SELECT COALESCE(SUM(total_amount), 0) INTO total
    FROM order_items
    WHERE order_id = order_id;
    RETURN total;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_organizations_updated_at
    BEFORE UPDATE ON organizations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_orders_updated_at
    BEFORE UPDATE ON orders
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_products_updated_at
    BEFORE UPDATE ON products
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_posts_updated_at
    BEFORE UPDATE ON posts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE INDEX idx_users_organization_id ON users(organization_id);

CREATE INDEX idx_users_status ON users(status);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);

CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

CREATE INDEX idx_products_category_id ON products(category_id);

CREATE INDEX idx_products_is_active ON products(is_active);

CREATE INDEX idx_product_variants_product_id ON product_variants(product_id);

CREATE INDEX idx_inventory_product_id ON inventory(product_id);

CREATE INDEX idx_inventory_warehouse_id ON inventory(warehouse_id);

CREATE INDEX idx_orders_user_id ON orders(user_id);

CREATE INDEX idx_orders_status ON orders(status);

CREATE INDEX idx_orders_created_at ON orders(created_at);

CREATE INDEX idx_order_items_order_id ON order_items(order_id);

CREATE INDEX idx_order_items_product_id ON order_items(product_id);

CREATE INDEX idx_payments_order_id ON payments(order_id);

CREATE INDEX idx_payments_status ON payments(status);

CREATE INDEX idx_posts_author_id ON posts(author_id);

CREATE INDEX idx_posts_status ON posts(status);

CREATE INDEX idx_comments_post_id ON comments(post_id);

CREATE INDEX idx_comments_author_id ON comments(author_id);

CREATE INDEX idx_notifications_user_id ON notifications(user_id);

CREATE INDEX idx_notifications_is_read ON notifications(is_read);

CREATE INDEX idx_audit_logs_organization_id ON audit_logs(organization_id);

CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);

CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);

ALTER TABLE users ENABLE ROW LEVEL SECURITY;

ALTER TABLE user_profiles ENABLE ROW LEVEL SECURITY;

ALTER TABLE orders ENABLE ROW LEVEL SECURITY;

ALTER TABLE order_items ENABLE ROW LEVEL SECURITY;

ALTER TABLE payments ENABLE ROW LEVEL SECURITY;

CREATE POLICY users_organization_policy ON users
    USING (organization_id = current_setting('app.current_organization_id', true)::bigint);

CREATE POLICY user_profiles_user_policy ON user_profiles
    USING (user_id = current_setting('app.current_user_id', true)::bigint);

CREATE POLICY orders_user_policy ON orders
    USING (user_id = current_setting('app.current_user_id', true)::bigint);

ALTER TABLE categories ADD CONSTRAINT categories_parent_fkey FOREIGN KEY (parent_id) REFERENCES categories(id);

ALTER TABLE departments ADD CONSTRAINT departments_parent_fkey FOREIGN KEY (parent_id) REFERENCES departments(id);

ALTER TABLE comments ADD CONSTRAINT comments_parent_fkey FOREIGN KEY (parent_id) REFERENCES comments(id);
