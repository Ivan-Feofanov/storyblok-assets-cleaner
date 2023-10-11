This CLI tool helps to bulk remove assets from storyblok
using the storyblok management API.

### Params:
 - api-url The storyblok management API url (default "https://mapi.storyblok.com/v1/")
 - api-token The storyblok management API token (required)
 - space-id The storyblok space id (required)
 - folder-id The storyblok folder id (required)

### Example
```bash
storyblok-bulk-remove-assets -api-token=xxx -space-id=123 -folder-id=456
```
