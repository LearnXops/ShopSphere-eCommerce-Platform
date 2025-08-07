-- Create notification service schema
-- This migration creates tables for notification templates, preferences, delivery tracking, and retry mechanisms

-- Create notification templates table
CREATE TABLE notification_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    channel VARCHAR(20) NOT NULL CHECK (channel IN ('email', 'sms', 'push')),
    subject VARCHAR(200),
    body_template TEXT NOT NULL,
    variables JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create notification preferences table
CREATE TABLE notification_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    channel VARCHAR(20) NOT NULL CHECK (channel IN ('email', 'sms', 'push')),
    category VARCHAR(50) NOT NULL,
    is_enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, channel, category)
);

-- Create notifications table for tracking sent notifications
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    template_id UUID REFERENCES notification_templates(id),
    channel VARCHAR(20) NOT NULL CHECK (channel IN ('email', 'sms', 'push')),
    recipient VARCHAR(255) NOT NULL,
    subject VARCHAR(200),
    body TEXT NOT NULL,
    variables JSONB DEFAULT '{}',
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'delivered', 'failed', 'bounced')),
    provider VARCHAR(50),
    provider_message_id VARCHAR(255),
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    scheduled_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    sent_at TIMESTAMP WITH TIME ZONE,
    delivered_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create notification delivery events table for tracking
CREATE TABLE notification_delivery_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    notification_id UUID NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
    event_type VARCHAR(20) NOT NULL CHECK (event_type IN ('sent', 'delivered', 'bounced', 'opened', 'clicked', 'unsubscribed')),
    event_data JSONB DEFAULT '{}',
    provider_event_id VARCHAR(255),
    occurred_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create notification retry queue table
CREATE TABLE notification_retry_queue (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    notification_id UUID NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
    retry_attempt INTEGER NOT NULL DEFAULT 1,
    scheduled_for TIMESTAMP WITH TIME ZONE NOT NULL,
    error_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for performance
CREATE INDEX idx_notification_templates_channel ON notification_templates(channel);
CREATE INDEX idx_notification_templates_name ON notification_templates(name);
CREATE INDEX idx_notification_preferences_user_id ON notification_preferences(user_id);
CREATE INDEX idx_notification_preferences_user_channel ON notification_preferences(user_id, channel);
CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_status ON notifications(status);
CREATE INDEX idx_notifications_channel ON notifications(channel);
CREATE INDEX idx_notifications_scheduled_at ON notifications(scheduled_at);
CREATE INDEX idx_notifications_created_at ON notifications(created_at);
CREATE INDEX idx_notification_delivery_events_notification_id ON notification_delivery_events(notification_id);
CREATE INDEX idx_notification_delivery_events_event_type ON notification_delivery_events(event_type);
CREATE INDEX idx_notification_retry_queue_scheduled_for ON notification_retry_queue(scheduled_for);
CREATE INDEX idx_notification_retry_queue_notification_id ON notification_retry_queue(notification_id);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_notification_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for updated_at
CREATE TRIGGER trigger_notification_templates_updated_at
    BEFORE UPDATE ON notification_templates
    FOR EACH ROW EXECUTE FUNCTION update_notification_updated_at();

CREATE TRIGGER trigger_notification_preferences_updated_at
    BEFORE UPDATE ON notification_preferences
    FOR EACH ROW EXECUTE FUNCTION update_notification_updated_at();

CREATE TRIGGER trigger_notifications_updated_at
    BEFORE UPDATE ON notifications
    FOR EACH ROW EXECUTE FUNCTION update_notification_updated_at();

-- Insert default notification templates
INSERT INTO notification_templates (name, channel, subject, body_template, variables) VALUES
-- Email templates
('welcome_email', 'email', 'Welcome to ShopSphere!', 
 'Hello {{.UserName}},\n\nWelcome to ShopSphere! We''re excited to have you join our community.\n\nBest regards,\nThe ShopSphere Team', 
 '{"UserName": "string"}'),
 
('order_confirmation_email', 'email', 'Order Confirmation - #{{.OrderNumber}}',
 'Hello {{.UserName}},\n\nThank you for your order! Your order #{{.OrderNumber}} has been confirmed.\n\nOrder Total: ${{.OrderTotal}}\n\nWe''ll send you updates as your order is processed.\n\nBest regards,\nThe ShopSphere Team',
 '{"UserName": "string", "OrderNumber": "string", "OrderTotal": "string"}'),

('order_shipped_email', 'email', 'Your Order Has Shipped - #{{.OrderNumber}}',
 'Hello {{.UserName}},\n\nGreat news! Your order #{{.OrderNumber}} has been shipped.\n\nTracking Number: {{.TrackingNumber}}\nCarrier: {{.Carrier}}\n\nYou can track your package using the tracking number above.\n\nBest regards,\nThe ShopSphere Team',
 '{"UserName": "string", "OrderNumber": "string", "TrackingNumber": "string", "Carrier": "string"}'),

('password_reset_email', 'email', 'Reset Your Password',
 'Hello {{.UserName}},\n\nYou requested to reset your password. Click the link below to reset it:\n\n{{.ResetLink}}\n\nThis link will expire in 1 hour.\n\nIf you didn''t request this, please ignore this email.\n\nBest regards,\nThe ShopSphere Team',
 '{"UserName": "string", "ResetLink": "string"}'),

-- SMS templates  
('order_confirmation_sms', 'sms', NULL,
 'ShopSphere: Your order #{{.OrderNumber}} has been confirmed. Total: ${{.OrderTotal}}. Thank you for shopping with us!',
 '{"OrderNumber": "string", "OrderTotal": "string"}'),

('order_shipped_sms', 'sms', NULL,
 'ShopSphere: Your order #{{.OrderNumber}} has shipped! Tracking: {{.TrackingNumber}} via {{.Carrier}}.',
 '{"OrderNumber": "string", "TrackingNumber": "string", "Carrier": "string"}'),

('delivery_notification_sms', 'sms', NULL,
 'ShopSphere: Your order #{{.OrderNumber}} has been delivered! Thank you for shopping with us.',
 '{"OrderNumber": "string"}'),

-- Push notification templates
('order_confirmation_push', 'push', 'Order Confirmed',
 'Your order #{{.OrderNumber}} has been confirmed. Total: ${{.OrderTotal}}',
 '{"OrderNumber": "string", "OrderTotal": "string"}'),

('order_shipped_push', 'push', 'Order Shipped',
 'Your order #{{.OrderNumber}} has shipped! Track it with {{.TrackingNumber}}',
 '{"OrderNumber": "string", "TrackingNumber": "string"}'),

('new_review_push', 'push', 'New Review',
 'Someone reviewed your product "{{.ProductName}}". Check it out!',
 '{"ProductName": "string"}');

-- Insert default notification preferences categories
-- These will be used as defaults when users sign up
INSERT INTO notification_preferences (user_id, channel, category, is_enabled) 
SELECT 
    '00000000-0000-0000-0000-000000000000'::uuid as user_id,
    channel,
    category,
    true as is_enabled
FROM (
    VALUES 
    ('email', 'order_updates'),
    ('email', 'marketing'),
    ('email', 'security'),
    ('email', 'reviews'),
    ('sms', 'order_updates'),
    ('sms', 'delivery'),
    ('push', 'order_updates'),
    ('push', 'reviews'),
    ('push', 'recommendations')
) AS defaults(channel, category);
