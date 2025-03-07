# Testing the Payment Flow with Stripe

This document provides instructions for testing the complete payment flow with Stripe, including webhook handling.

## Prerequisites

1. A Stripe account with test mode enabled
2. The Stripe CLI installed for webhook testing
3. Environment variables set up with your Stripe test keys

## Setting Up Stripe CLI for Webhook Testing

1. Install the Stripe CLI following the instructions at https://stripe.com/docs/stripe-cli

2. Log in to your Stripe account:
   ```
   stripe login
   ```

3. Start forwarding webhook events to your local server:
   ```
   stripe listen --forward-to http://localhost:3000/webhook
   ```

   This will output a webhook signing secret. Copy this secret and set it as your `STRIPE_WEBHOOK_SECRET` environment variable.

## Testing the Complete Payment Flow

### 1. Create a Test User

1. Register a new user on your application
2. Log in with the new user

### 2. Subscribe to a Plan

1. Navigate to the pricing page (`/pricing`)
2. Select a subscription plan (e.g., monthly)
3. Click the "Subscribe" button
4. You'll be redirected to the Stripe Checkout page
5. Use a test card number:
   - For successful payments: `4242 4242 4242 4242`
   - For failed payments: `4000 0000 0000 0002`
6. Enter any future expiration date, any 3-digit CVC, and any billing information
7. Complete the checkout process

### 3. Verify Webhook Events

When you complete a checkout, Stripe will send webhook events to your application. With the Stripe CLI forwarding enabled, you should see these events in your terminal and in your application logs.

Key events to look for:
- `checkout.session.completed` - Sent when the checkout is completed
- `customer.subscription.created` - Sent when a new subscription is created
- `invoice.paid` - Sent when an invoice is paid

### 4. Verify User Subscription Status

1. After completing the checkout, navigate to your user profile or dashboard
2. Verify that your subscription status is updated correctly
3. Verify that you can access premium features (e.g., create more than 2 guns)

### 5. Testing Subscription Updates and Cancellations

You can simulate subscription updates and cancellations using the Stripe CLI:

1. To simulate a subscription update:
   ```
   stripe trigger customer.subscription.updated
   ```

2. To simulate a subscription cancellation:
   ```
   stripe trigger customer.subscription.deleted
   ```

3. Verify that your application handles these events correctly by checking the user's subscription status

## Monitoring Webhook Health

The application includes a webhook monitoring system that tracks the health of webhook events. You can access this information at:

```
/admin/webhook-health
```

This endpoint requires admin privileges and provides information about:
- Total webhook requests received
- Success and failure rates
- Timestamps of the last request and last error
- Details of any errors

## Troubleshooting

### Common Issues

1. **Webhook Verification Failures**
   - Ensure your `STRIPE_WEBHOOK_SECRET` environment variable is set correctly
   - Check that you're using the correct webhook secret from the Stripe CLI

2. **Missing Webhook Events**
   - Verify that the Stripe CLI is running and forwarding events
   - Check your application logs for any errors in webhook processing

3. **Subscription Status Not Updating**
   - Check your application logs for errors in webhook processing
   - Verify that the user ID in the webhook event matches the expected user

### Viewing Webhook Events in Stripe Dashboard

You can view all webhook events in the Stripe Dashboard:
1. Go to the Stripe Dashboard
2. Navigate to Developers > Events
3. You'll see a list of all events, including those sent to your webhook endpoint

## Production Considerations

When moving to production:

1. Update your environment variables with production Stripe keys
2. Set `APP_ENV=production` in your environment
3. Update your webhook endpoint URL in the Stripe Dashboard
4. Implement proper error handling and monitoring for webhook failures
5. Consider setting up redundancy for critical webhook events

Remember that in production, Stripe will send real webhook events to your endpoint, so ensure your server is properly secured and can handle the expected load. 