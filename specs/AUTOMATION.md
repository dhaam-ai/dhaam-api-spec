# API automation guide

This guide is for QA automation against the DhaamAI public API, imported into Apidog.
It covers the customer-app flows and the admin setup calls used to create test data.
Every host, token and tenant id is a placeholder. Real values belong in the Apidog environment,
never in this repository.

## Files

| File | What it is | Import into Apidog |
|------|------------|--------------------|
| `specs/openapi.yml` | Merged spec, all in-scope REST operations | Import, OpenAPI/Swagger, upload the file |
| `specs/auth.yml` | dh-auth: check-user, login, refresh, social callback | Same, or per service |
| `specs/user_profile.yml` | dh-user-profile: customer OTP and login, profile, addresses, store login and info | Same |
| `specs/profile.yml` | dh-profile: store creation and outlet setup | Same |
| `specs/catalog.yml` | dh-catalog-service: admin product setup, item list | Same |
| `specs/menu.yml` | dh-menu: admin menu setup | Same |
| `specs/order.yml` | dh-order-api: quote, create, list, detail, cancel, payment modes, customer payment calls | Same |
| `specs/gateway.yml` | dh-gateway-api: storefront GraphQL, POST `/api/v1/storefront/graphql` | Same |
| `specs/gateway.graphql` | Served GraphQL schema for the gateway | Import, GraphQL, upload the schema file |

Per-service files are the source. `specs/openapi.yml` is generated from them by `merge_specs.go`
(`make merge`). Do not edit the merged file by hand.

## Environments (Apidog)

Create one environment per target. Define these variables.

| Variable | Value | Notes |
|----------|-------|-------|
| `baseUrl` | `https://<public-host>` | The console or shop host for the target environment. Set in Apidog only. |
| `token` | Bearer token | Set by the login step below. Never commit. |
| `tenantId` | Tenant id or uuid used by the test account | [TK: confirm which identifier the test tenant uses; see "Known gaps"]. |
| `customerPhone`, `customerOtp` | Test customer identifiers | Held in the Apidog environment only. |

Every path in the specs is a full public path, for example `/api/v1/orders/quote`. The server
in each file is `{{baseUrl}}`, so requests resolve against the environment's `baseUrl`.

## Authentication

- Most customer and admin operations need `Authorization: Bearer {{token}}`.
- Public operations (no token) are the login and check steps, the social callback, the store
  login, and the storefront GraphQL endpoint. Each operation's `security` field says which.
- Send `Authorization: Bearer {{token}}` on operations that need it.
- A 401 means the token is missing or expired. Refresh it, then retry once.

### Getting a token

1. `POST /api/v1/auth/check-user` (public). Confirm the identifier is known before login.
2. Customer sign-in (public steps):
   1. `POST /api/v1/users/otp/initiate` with the customer identifier.
   2. `POST /api/v1/users/otp/verify/{otptype}` with the code.
   3. Or password login: `POST /api/v1/users/customer/login`.
   [TK: the initiate response does not match what the customer app reads (see "Known gaps").
   Confirm the verify step returns a token and whether the test environment accepts a fixed
   OTP. Do not write a fixed OTP into this repository.]
3. Merchant or store sign-in (public): `POST /api/v1/users/store/login` (password login,
   dh-user-profile). Use the returned token as `token` for the admin setup calls.
   [TK: lead to confirm this is the admin setup account type. The dh-auth
   `POST /api/v1/auth/login` is a separate merchant login.]
4. Renew: `POST /api/v1/auth/refresh` with the refresh token. Use it when a call returns 401.

Rate limits apply to the OTP, login and reset calls. Defaults documented in the specs
(per `RATE_*` settings, which the environment can override):

- OTP initiate and verify, login, reset: 10 per hour per IP, and 5 per hour per identifier on
  staging and production (10 per hour per identifier elsewhere). Block is 1 hour on staging and
  production, 1 minute elsewhere.
- Signup: 10 per hour per IP.
- Social callback: 20 per 15 minutes per IP, blocked for 1 hour after that. 100 per 15 minutes per
  tenant, keyed on the request's `tenant_id`.

A rate-limited response is HTTP 429 with a body that is not the standard envelope:

```json
{"error": "rate_limit_exceeded", "message": "Too many requests. Please try again later.", "retry_after": 60}
```

It also sends `Retry-After` and `X-RateLimit-*` headers. Wait `retry_after` seconds, or use a
fresh identifier in the test data.

## Response and error format

Success:

```json
{"success": true, "data": {}, "meta": {}}
```

Error:

```json
{
  "success": false,
  "error": {"type": "VALIDATION_FAILED", "message": "...", "details": {}},
  "meta": {"timestamp": "2026-01-01T00:00:00Z", "request_id": "..."}
}
```

- Validation failures are HTTP 400 with type `VALIDATION_FAILED`. Field errors are in
  `error.details.fields[]`, each with `field`, `message` and `value`.
- An unexpected server error is HTTP 500 with type `INTERNAL_ERROR` and the message
  "An internal error occurred".
- The envelope shape comes from the shared response library. The services pin a newer version
  than the one this was checked against. [TK: re-check the envelope against the pinned version
  before relying on exact type names.]
- Some services use other error bodies. Menu returns `{message}` for framework errors. Check the
  response schema of each operation.

## Pagination

- Catalog item list: `POST /api/v1/catalog/item/list` takes `limit`, `current_page`,
  `previous_page` and `filters`.
- Order list: `POST /api/v1/orders/list`. [TK: confirm page and size field names in the order
  spec before relying on them.]

## Ordering of calls

Customer order flow:

1. Sign in (see Getting a token).
2. Browse the storefront through `POST /api/v1/storefront/graphql`. The examples in `gateway.yml`
   are the operations the customer app sends, with synthetic variables.
3. Quote: `POST /api/v1/orders/quote`. Keep the quote id for the next step.
4. Create: `POST /api/v1/orders/create` with the quote. Quotes are required before create.
5. Read: `GET /api/v1/orders/{id}` for detail, `POST /api/v1/orders/list` for the list.
6. Cancel: `GET /api/v1/orders/{id}/cancellation-preview` first, then `POST /api/v1/orders/{id}`
   with `status` 9 (Cancelled) and a reason.

### Place an order (quote, then create)

1. **Quote.** `POST /api/v1/orders/quote` prices the cart for one outlet and returns a quote id.
   The spec has three ready examples: delivery now with cash, scheduled delivery, pickup with card.

   ```json
   {
     "customer_id": "{{customerId}}",
     "outlet_id": "{{outletId}}",
     "address": "Synthetic Street 1, Testville",
     "latitude": 12.9716,
     "longitude": 77.5946,
     "delivery_mode": 1,
     "order_mode": 2,
     "payment_via": 1,
     "timezone": "Asia/Kolkata",
     "items": [
       { "item_id": "{{itemId}}", "menu_id": "{{menuId}}", "entity_type": 1, "quantity": 2,
         "unit_price": 250.00, "notes": "" }
     ]
   }
   ```

   - `delivery_mode` 1 = home delivery, 2 = pickup (see the value reference below).
   - `order_mode` 2 = now; 1 = scheduled, and then also send `scheduled_for` (ISO 8601 time).
   - `payment_via` 1 = cash on delivery, 2 = card, 3 = wallet. Optional on the quote, required
     before create.
   - Quote again after any cart change. A quote lives 30 minutes and is used up by the create.
2. **Create.** `POST /api/v1/orders/create` with the quote id from step 1:

   ```json
   {
     "quotation_id": "{{quotationId}}",
     "customer_id": "{{customerId}}",
     "client_request_id": "test-run-1-order-0001"
   }
   ```

   - `client_request_id` must be new for every order. Sending the same one again returns the
     first order instead of creating a second, so a retry after a timeout is safe.
   - An expired or used quote returns 404 `QUOTE_NOT_FOUND`: quote again, then create.
3. **Check it.** `GET /api/v1/orders/{id}`: `order_status` starts at 1 (New).

## Value reference

Every code a request or response uses, in one place. Each field also carries its `enum` and an
example in the spec.

| Field | Values |
|-------|--------|
| `role` (login, refresh, social) | `"1"` admin, `"2"` merchant (store owner), `"3"` manager, `"4"` customer, `"5"` super admin, `"6"` agent, `"101"` outlet user. Strings. |
| Social login `provider` (URL) | `google`, `apple`, `facebook`. The body's `provider_token` is that provider's token: Google id_token, Apple id_token (plus `nonce`), Facebook access token. |
| Social login `action` | `callback` (sign in with a token the app already has), `initiate` (start the provider's redirect flow) |
| OTP `type` | `signup`, `login`, `password_change`, `email_change`, `phone_change`, `enable2fa` |
| OTP `provider_type` | `ses` (email), `sms` (text), `whatsapp` (phone) |
| `identifier_type` | `email`, `phone` |
| `platform` | `web`, `android`, `ios` |
| `device_type` (user profile) | `web`, `ios`, `android`, `desktop` |
| `verification_status` | 0 not verified, 1 email, 2 phone, 3 email and phone, 4 verified by an admin |
| `business_type` | Ids from the storefront config query `businessTypes`, for example 1 restaurant, 2 grocery, 3 cafe |
| `delivery_mode` (orders) | 1 home delivery, 2 pickup, 3 takeaway, 4 dine-in. Ids from the config query `deliveryModes`. |
| `order_mode` | 1 scheduled (send `scheduled_for`), 2 now |
| `payment_via`, `payment_mode`, `gateway_type` | 1 cash on delivery, 2 card, 3 wallet |
| `order_status` | 1 New, 2 Accepted, 3 Assigned, 4 Processed, 5 Order ready, 6 Pickup, 7 Dispatched, 8 Delivered, 9 Cancelled, 10 Deleted, 11 Pickup started |
| Cancel `status` | 9 |
| `refund_method` | `Source` (back to the original payment), `Wallet` |
| `entity_type` (items) | 1 product, 2 modifier, 3 modifier group, 4 upsell, 5 deal, 6 variant, 7 bundle |
| Category `category_type` | 1 business category (platform), 2 product category (a store's menu); store setup uses 2 |
| Category `type` | `parent` (top level), `child` (subcategory) |
| Menu create `{status}` (URL) | `0` inactive, `1` active |
| `day_of_week` (menu and outlet hours) | 1 Monday … 7 Sunday |
| Outlet hours `delivery_mode` | 1 delivery hours, 2 pickup hours |
| `schedule_type` | 1 regular opening hours |

Admin setup, to create a store that accepts orders. Run these in order. [TK: confirm the order
and which calls are required; see the mapping below.]

1. Sign in as admin (see Getting a token).
2. Create the store: `POST /api/v1/users/store/register` (no token). The store can log in afterwards with
   the same credentials, using `POST /api/v1/users/store/login`.
3. Set the store basics: `PUT /api/v1/users/store/info` (name, description, address, business
   type, categories, logo).
4. Create the outlet and its hours: `POST /api/v1/profile/outlets`, then
   `PUT /api/v1/profile/outlets/{id}` to change it. [TK: the new outlet id is not returned, and the
   schedule enum values are not confirmed.]
5. Create products: `POST /api/v1/catalog/item`, then read and update with
   `GET` and `PUT /api/v1/catalog/item/{id}`. List with `POST /api/v1/catalog/item/list`.
6. Create a menu: `POST /api/v1/catalog/menu/create-menu/{status}` (status 0 or 1).
7. Set the menu hours: `POST /api/v1/catalog/menu/schedule/{menu_id}`, change them with
   `PUT /api/v1/catalog/menu/schedule/{id}`.
8. Read the hours (timeslots):
   - Menu hours: `GET /api/v1/catalog/menu/schedule/menu/{id}` (the response carries `time_slots`),
     `GET /api/v1/catalog/menu/schedule/{id}`, or the list `GET /api/v1/catalog/menu/schedule/`.
   - Outlet hours: `GET /api/v1/users/outlets/{id}` returns the outlet with its delivery and pickup
     schedules. Outlet hours are set in step 4 (`delivery_schedules`, `pickup_schedules`).

There are two kinds of opening hours, and a store takes orders only when both allow it: quotes and
orders check that the outlet is open (outlet hours) and that the menu is inside its schedule (menu
hours). For an all-day test store, set both to cover the whole day, every day.

### Admin setup mapping

| Setup group | Endpoint | Spec and operationId | Status |
|-------------|----------|----------------------|--------|
| 1. Store creation | `POST /api/v1/users/store/register`, then outlet add `POST /api/v1/profile/outlets` as the follow-up step | `user_profile.yml`, `users_registerStore`; `profile.yml`, `profile_addOutlet` | Store can log in afterwards. |
| 2. Timeslots (opening hours) | Menu hours: `GET /api/v1/catalog/menu/schedule/menu/{id}`, `GET /api/v1/catalog/menu/schedule/{id}`, `GET /api/v1/catalog/menu/schedule/`, `POST` and `PUT /api/v1/catalog/menu/schedule/{id}`. Outlet hours: `GET /api/v1/users/outlets/{id}`; set with the outlet add and update calls | `menu.yml`, `menu_getMenuScheduleByMenuId`, `menu_getMenuSchedule`, `menu_listMenuSchedules`, `menu_createMenuSchedule`, `menu_updateMenuSchedule`; `user_profile.yml`, `users_getOutletDetail`; `profile.yml`, `profile_addOutlet`, `profile_updateOutlet` | confirmed in code |
| 3. Add menu | `POST /api/v1/catalog/menu/create-menu/{status}` | `menu.yml`, `menu_createMenu` | confirmed in code |
| 4. Products | `POST /api/v1/catalog/item`, `GET` and `PUT /api/v1/catalog/item/{id}`, `POST /api/v1/catalog/item/list` | `catalog.yml`, `catalog_createItem`, `catalog_getItem`, `catalog_updateItem`, `catalog_listItems` | confirmed in code |
| 5. Basic store config | `PUT /api/v1/users/store/info` and `POST /api/v1/profile/outlets`, `PUT /api/v1/profile/outlets/{id}` | `user_profile.yml`, `users_updateStoreInfo`; `profile.yml`, `profile_addOutlet`, `profile_updateOutlet` | Basic fields only. Other outlet body fields are out of scope. |

Merchant login for admin setup: `POST /api/v1/users/store/login` (`users_storeLogin`). See
"Getting a token".

## Customer storefront (GraphQL)

- Endpoint: `POST /api/v1/storefront/graphql`. Public, no token.
- Body: `{"query": "...", "operationName": "...", "variables": {...}}`.
- Response: `{"data": {...}, "errors": [...]}`. A GraphQL error can come with HTTP 200.
- The 17 named examples in `gateway.yml` are the app's own operations.

## Known gaps and failures the suite will hit

These are real disagreements found in the code and the app. Each is also an `x-todo` in its spec.

- `storeSlots`, `itemUpsells` and `promoMenu` fail validation against the served schema. The app
  selects fields the schema does not define. Expect GraphQL errors and no data.
- `itemUpsells` sends numeric item ids from the app, but the schema types them as strings.
- Customer OTP initiate: the app reads `data.session` and `data.username`, but the handler returns
  only `type` and `login_methods`. The app's verify step cannot proceed from that response.
- Confirm whether this operation needs a bearer token.
- Wrong credentials and expired refresh tokens return HTTP 500 in the auth service, not 401.
- Check-user with a malformed identifier returns HTTP 500.
- Cancelling an order past dispatch returns HTTP 500.
- The served GraphQL schema declares two `MenuCategory` fields twice. Strict tools may reject
  `specs/gateway.graphql` on import.

## Importing into Apidog

1. Create the environment with `baseUrl` and the other variables above.
2. Import the OpenAPI file: Import, then OpenAPI/Swagger, then upload `specs/openapi.yml`.
3. For GraphQL requests: create a GraphQL request against `POST /api/v1/storefront/graphql` and
   import `specs/gateway.graphql` as the schema.
4. Import each per-service file only if you need them separately. The merged file already holds
   all of them.

The import has not been run in Apidog yet. [TK: record the Apidog version and any import errors.]
