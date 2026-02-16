-- Seed a customer first to have a valid customer_id
INSERT INTO customers (id, name, email, phone, address)
VALUES ('00000000-0000-0000-0000-000000000001', 'Admin Customer', 'admin_customer@example.com', '0123456789', 'Admin Address')
ON CONFLICT (id) DO NOTHING;

-- Seed 100 notes for the admin customer
DO $$
BEGIN
    FOR i IN 1..100 LOOP
        INSERT INTO notes (content, customer_id)
        VALUES ('Note ' || i || ' for admin customer', '00000000-0000-0000-0000-000000000001');
    END LOOP;
END $$;
