// MongoDB initialization script for ShopSphere eCommerce platform
// This script creates collections and indexes for product catalog and analytics

// Switch to the shopsphere database
use('shopsphere');

// Product Catalog Collections
// ===========================

// Product catalog collection (rich product data)
db.createCollection('product_catalog', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      required: ['product_id', 'name', 'status'],
      properties: {
        product_id: {
          bsonType: 'string',
          description: 'Product ID from PostgreSQL'
        },
        name: {
          bsonType: 'string',
          description: 'Product name'
        },
        description: {
          bsonType: 'string',
          description: 'Rich product description with HTML'
        },
        short_description: {
          bsonType: 'string',
          description: 'Short product description'
        },
        specifications: {
          bsonType: 'object',
          description: 'Detailed product specifications'
        },
        features: {
          bsonType: 'array',
          items: {
            bsonType: 'string'
          },
          description: 'Product features list'
        },
        media: {
          bsonType: 'object',
          properties: {
            images: {
              bsonType: 'array',
              items: {
                bsonType: 'object',
                properties: {
                  url: { bsonType: 'string' },
                  alt_text: { bsonType: 'string' },
                  sort_order: { bsonType: 'int' },
                  is_primary: { bsonType: 'bool' }
                }
              }
            },
            videos: {
              bsonType: 'array',
              items: {
                bsonType: 'object',
                properties: {
                  url: { bsonType: 'string' },
                  title: { bsonType: 'string' },
                  duration: { bsonType: 'int' }
                }
              }
            }
          }
        },
        seo: {
          bsonType: 'object',
          properties: {
            meta_title: { bsonType: 'string' },
            meta_description: { bsonType: 'string' },
            meta_keywords: { bsonType: 'array' },
            slug: { bsonType: 'string' }
          }
        },
        status: {
          bsonType: 'string',
          enum: ['active', 'inactive', 'discontinued']
        },
        created_at: {
          bsonType: 'date'
        },
        updated_at: {
          bsonType: 'date'
        }
      }
    }
  }
});

// Create indexes for product catalog
db.product_catalog.createIndex({ 'product_id': 1 }, { unique: true });
db.product_catalog.createIndex({ 'name': 'text', 'description': 'text', 'features': 'text' });
db.product_catalog.createIndex({ 'status': 1 });
db.product_catalog.createIndex({ 'seo.slug': 1 });
db.product_catalog.createIndex({ 'created_at': 1 });
db.product_catalog.createIndex({ 'updated_at': 1 });

// Category hierarchy collection (for complex category structures)
db.createCollection('category_hierarchy', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      required: ['category_id', 'name', 'path'],
      properties: {
        category_id: {
          bsonType: 'string',
          description: 'Category ID from PostgreSQL'
        },
        name: {
          bsonType: 'string'
        },
        description: {
          bsonType: 'string'
        },
        path: {
          bsonType: 'string',
          description: 'Materialized path for hierarchy'
        },
        ancestors: {
          bsonType: 'array',
          items: {
            bsonType: 'string'
          },
          description: 'Array of ancestor category IDs'
        },
        children: {
          bsonType: 'array',
          items: {
            bsonType: 'string'
          },
          description: 'Array of direct child category IDs'
        },
        level: {
          bsonType: 'int'
        },
        sort_order: {
          bsonType: 'int'
        },
        is_active: {
          bsonType: 'bool'
        },
        metadata: {
          bsonType: 'object',
          description: 'Additional category metadata'
        }
      }
    }
  }
});

db.category_hierarchy.createIndex({ 'category_id': 1 }, { unique: true });
db.category_hierarchy.createIndex({ 'path': 1 });
db.category_hierarchy.createIndex({ 'ancestors': 1 });
db.category_hierarchy.createIndex({ 'level': 1 });
db.category_hierarchy.createIndex({ 'is_active': 1 });

// Analytics Collections
// =====================

// User behavior analytics
db.createCollection('user_analytics', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      required: ['user_id', 'session_id', 'event_type', 'timestamp'],
      properties: {
        user_id: {
          bsonType: 'string',
          description: 'User ID (null for anonymous users)'
        },
        session_id: {
          bsonType: 'string',
          description: 'Session identifier'
        },
        event_type: {
          bsonType: 'string',
          enum: ['page_view', 'product_view', 'add_to_cart', 'remove_from_cart', 'search', 'purchase', 'login', 'logout']
        },
        event_data: {
          bsonType: 'object',
          description: 'Event-specific data'
        },
        page_url: {
          bsonType: 'string'
        },
        referrer: {
          bsonType: 'string'
        },
        user_agent: {
          bsonType: 'string'
        },
        ip_address: {
          bsonType: 'string'
        },
        device_info: {
          bsonType: 'object',
          properties: {
            device_type: { bsonType: 'string' },
            browser: { bsonType: 'string' },
            os: { bsonType: 'string' },
            screen_resolution: { bsonType: 'string' }
          }
        },
        timestamp: {
          bsonType: 'date'
        }
      }
    }
  }
});

db.user_analytics.createIndex({ 'user_id': 1 });
db.user_analytics.createIndex({ 'session_id': 1 });
db.user_analytics.createIndex({ 'event_type': 1 });
db.user_analytics.createIndex({ 'timestamp': 1 });
db.user_analytics.createIndex({ 'timestamp': 1, 'event_type': 1 });

// Product analytics
db.createCollection('product_analytics', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      required: ['product_id', 'date'],
      properties: {
        product_id: {
          bsonType: 'string'
        },
        date: {
          bsonType: 'date',
          description: 'Date for daily aggregation'
        },
        metrics: {
          bsonType: 'object',
          properties: {
            views: { bsonType: 'int' },
            unique_views: { bsonType: 'int' },
            add_to_cart: { bsonType: 'int' },
            purchases: { bsonType: 'int' },
            revenue: { bsonType: 'double' },
            conversion_rate: { bsonType: 'double' },
            bounce_rate: { bsonType: 'double' },
            avg_time_on_page: { bsonType: 'double' }
          }
        },
        updated_at: {
          bsonType: 'date'
        }
      }
    }
  }
});

db.product_analytics.createIndex({ 'product_id': 1, 'date': 1 }, { unique: true });
db.product_analytics.createIndex({ 'date': 1 });
db.product_analytics.createIndex({ 'metrics.views': -1 });
db.product_analytics.createIndex({ 'metrics.purchases': -1 });

// Search analytics
db.createCollection('search_analytics', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      required: ['query', 'timestamp'],
      properties: {
        query: {
          bsonType: 'string',
          description: 'Search query'
        },
        normalized_query: {
          bsonType: 'string',
          description: 'Normalized search query for aggregation'
        },
        user_id: {
          bsonType: 'string'
        },
        session_id: {
          bsonType: 'string'
        },
        results_count: {
          bsonType: 'int'
        },
        clicked_results: {
          bsonType: 'array',
          items: {
            bsonType: 'object',
            properties: {
              product_id: { bsonType: 'string' },
              position: { bsonType: 'int' },
              clicked_at: { bsonType: 'date' }
            }
          }
        },
        filters_applied: {
          bsonType: 'object',
          description: 'Filters applied to the search'
        },
        sort_order: {
          bsonType: 'string'
        },
        timestamp: {
          bsonType: 'date'
        }
      }
    }
  }
});

db.search_analytics.createIndex({ 'query': 1 });
db.search_analytics.createIndex({ 'normalized_query': 1 });
db.search_analytics.createIndex({ 'timestamp': 1 });
db.search_analytics.createIndex({ 'user_id': 1 });

// Recommendation data
db.createCollection('user_preferences', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      required: ['user_id'],
      properties: {
        user_id: {
          bsonType: 'string'
        },
        categories: {
          bsonType: 'object',
          description: 'Category preferences with scores'
        },
        brands: {
          bsonType: 'object',
          description: 'Brand preferences with scores'
        },
        price_range: {
          bsonType: 'object',
          properties: {
            min: { bsonType: 'double' },
            max: { bsonType: 'double' }
          }
        },
        viewed_products: {
          bsonType: 'array',
          items: {
            bsonType: 'object',
            properties: {
              product_id: { bsonType: 'string' },
              score: { bsonType: 'double' },
              last_viewed: { bsonType: 'date' }
            }
          }
        },
        purchased_products: {
          bsonType: 'array',
          items: {
            bsonType: 'object',
            properties: {
              product_id: { bsonType: 'string' },
              score: { bsonType: 'double' },
              purchased_at: { bsonType: 'date' }
            }
          }
        },
        updated_at: {
          bsonType: 'date'
        }
      }
    }
  }
});

db.user_preferences.createIndex({ 'user_id': 1 }, { unique: true });
db.user_preferences.createIndex({ 'updated_at': 1 });

// Product recommendations cache
db.createCollection('product_recommendations', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      required: ['product_id', 'recommendations'],
      properties: {
        product_id: {
          bsonType: 'string'
        },
        recommendations: {
          bsonType: 'array',
          items: {
            bsonType: 'object',
            properties: {
              product_id: { bsonType: 'string' },
              score: { bsonType: 'double' },
              reason: { bsonType: 'string' }
            }
          }
        },
        algorithm_version: {
          bsonType: 'string'
        },
        generated_at: {
          bsonType: 'date'
        },
        expires_at: {
          bsonType: 'date'
        }
      }
    }
  }
});

db.product_recommendations.createIndex({ 'product_id': 1 }, { unique: true });
db.product_recommendations.createIndex({ 'expires_at': 1 });

// Inventory snapshots for analytics
db.createCollection('inventory_snapshots', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      required: ['date', 'products'],
      properties: {
        date: {
          bsonType: 'date',
          description: 'Snapshot date'
        },
        products: {
          bsonType: 'array',
          items: {
            bsonType: 'object',
            properties: {
              product_id: { bsonType: 'string' },
              sku: { bsonType: 'string' },
              stock: { bsonType: 'int' },
              reserved_stock: { bsonType: 'int' },
              price: { bsonType: 'double' },
              status: { bsonType: 'string' }
            }
          }
        },
        created_at: {
          bsonType: 'date'
        }
      }
    }
  }
});

db.inventory_snapshots.createIndex({ 'date': 1 }, { unique: true });

print('MongoDB collections and indexes created successfully!');
print('Collections created:');
print('- product_catalog');
print('- category_hierarchy');
print('- user_analytics');
print('- product_analytics');
print('- search_analytics');
print('- user_preferences');
print('- product_recommendations');
print('- inventory_snapshots');