-- Insert default system categories
INSERT INTO categories (name, type, icon, color, is_system) VALUES
-- Income categories
('Зарплата', 'income', '💰', '#4CAF50', true),
('Фриланс', 'income', '💻', '#8BC34A', true),
('Инвестиции', 'income', '📈', '#00BCD4', true),
('Подарки', 'income', '🎁', '#E91E63', true),
('Другие доходы', 'income', '💵', '#9C27B0', true),
-- Expense categories
('Продукты', 'expense', '🛒', '#FF5722', true),
('Транспорт', 'expense', '🚗', '#795548', true),
('Жилье', 'expense', '🏠', '#607D8B', true),
('Развлечения', 'expense', '🎮', '#FF9800', true),
('Здоровье', 'expense', '💊', '#F44336', true),
('Одежда', 'expense', '👕', '#3F51B5', true),
('Образование', 'expense', '📚', '#009688', true),
('Рестораны', 'expense', '🍽️', '#FFC107', true),
('Коммунальные услуги', 'expense', '💡', '#9E9E9E', true),
('Другие расходы', 'expense', '📦', '#673AB7', true)
ON CONFLICT DO NOTHING;