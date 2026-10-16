-- Development Only Seed Script
-- Usage: psql -U postgres -d employee_db -f database/seed.sql

INSERT INTO employees (name, email, department, designation, joining_date, status)
VALUES
    ('Arjun Nair', 'arjun.nair@beinex.com', 'Engineering', 'Lead Cloud Architect', '2023-01-10', 'Active'),
    ('Priya Sharma', 'priya.sharma@beinex.com', 'Consulting', 'Senior Data Consultant', '2023-04-15', 'Active'),
    ('Rahul Menon', 'rahul.menon@beinex.com', 'Engineering', 'Senior Full Stack Engineer', '2023-08-01', 'Active'),
    ('Sneha Patel', 'sneha.patel@beinex.com', 'Human Resources', 'HR Business Partner', '2022-11-20', 'Active'),
    ('Anand Krishnan', 'anand.k@beinex.com', 'Finance', 'Financial Analyst', '2024-02-01', 'Inactive'),
    ('Fathima Zahra', 'fathima.z@beinex.com', 'Marketing', 'Digital Marketing Specialist', '2024-03-12', 'Active'),
    ('Vikram Das', 'vikram.das@beinex.com', 'Operations', 'Operations Manager', '2022-06-05', 'Active'),
    ('Deepa Varma', 'deepa.varma@beinex.com', 'Consulting', 'Business Intelligence Analyst', '2024-01-18', 'Active')
ON CONFLICT (LOWER(email)) DO NOTHING;
