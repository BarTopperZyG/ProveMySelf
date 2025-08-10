import { FC } from 'react';
import { BasePropertyEditor } from './base-property-editor';
import { CanvasItem } from '@/types/quiz';
import { RJSFSchema, UiSchema } from '@rjsf/utils';

interface OrderingEditorProps {
  item: CanvasItem;
  className?: string;
}

const orderingSchema: RJSFSchema = {
  type: 'object',
  title: 'Ordering Question Properties',
  properties: {
    title: {
      type: 'string',
      title: 'Question Text',
      description: 'Instructions for the ordering task',
      minLength: 1,
      maxLength: 500
    },
    required: {
      type: 'boolean',
      title: 'Required Question',
      description: 'Users must complete the ordering to proceed',
      default: false
    },
    points: {
      type: 'number',
      title: 'Points',
      description: 'Points awarded for correct ordering',
      minimum: 0,
      maximum: 1000
    },
    explanation: {
      type: 'string',
      title: 'Explanation',
      description: 'Feedback explaining the correct order (optional)',
      maxLength: 1000
    },
    content: {
      type: 'object',
      title: 'Ordering Configuration',
      properties: {
        items: {
          type: 'array',
          title: 'Items to Order',
          description: 'List of items that users need to put in correct order',
          minItems: 2,
          maxItems: 10,
          items: {
            type: 'object',
            title: 'Ordering Item',
            properties: {
              id: {
                type: 'string',
                title: 'Item ID',
                description: 'Unique identifier for this item',
                default: () => `item_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
              },
              text: {
                type: 'string',
                title: 'Item Text',
                description: 'The text displayed for this item',
                minLength: 1,
                maxLength: 500
              },
              correctOrder: {
                type: 'number',
                title: 'Correct Position',
                description: 'The correct position for this item (1 = first)',
                minimum: 1,
                maximum: 10
              }
            },
            required: ['id', 'text', 'correctOrder']
          },
          default: [
            { id: 'item_1', text: 'First item in sequence', correctOrder: 1 },
            { id: 'item_2', text: 'Second item in sequence', correctOrder: 2 },
            { id: 'item_3', text: 'Third item in sequence', correctOrder: 3 },
            { id: 'item_4', text: 'Fourth item in sequence', correctOrder: 4 }
          ]
        },
        partialCredit: {
          type: 'boolean',
          title: 'Allow Partial Credit',
          description: 'Award partial points for partially correct ordering',
          default: true
        },
        shuffleItems: {
          type: 'boolean',
          title: 'Shuffle Item Order',
          description: 'Present items in random order to users',
          default: true
        },
        showNumbers: {
          type: 'boolean',
          title: 'Show Position Numbers',
          description: 'Display numbers next to items during ordering',
          default: false
        },
        orientation: {
          type: 'string',
          title: 'Layout Orientation',
          description: 'How to display the ordering items',
          enum: ['vertical', 'horizontal'],
          enumNames: ['Vertical (stacked)', 'Horizontal (side by side)'],
          default: 'vertical'
        }
      },
      required: ['items']
    }
  },
  required: ['title', 'required', 'content']
};

const orderingUiSchema: UiSchema = {
  title: {
    'ui:widget': 'textarea',
    'ui:options': {
      rows: 2
    },
    'ui:placeholder': 'e.g., Arrange these steps in the correct order...'
  },
  explanation: {
    'ui:widget': 'textarea',
    'ui:options': {
      rows: 3
    },
    'ui:placeholder': 'Explain why this order is correct...'
  },
  content: {
    'ui:title': 'Ordering Settings',
    'ui:order': ['items', 'partialCredit', 'shuffleItems', 'showNumbers', 'orientation'],
    items: {
      'ui:options': {
        addable: true,
        removable: true,
        orderable: false // We want manual control over correctOrder
      },
      items: {
        'ui:order': ['text', 'correctOrder', 'id'],
        id: {
          'ui:widget': 'hidden'
        },
        text: {
          'ui:placeholder': 'Enter item text...',
          'ui:widget': 'textarea',
          'ui:options': {
            rows: 2
          }
        },
        correctOrder: {
          'ui:help': 'The position this item should be in when correctly ordered (1 = first, 2 = second, etc.)'
        }
      }
    },
    partialCredit: {
      'ui:help': 'When enabled, users get partial points based on how many items they place correctly'
    },
    shuffleItems: {
      'ui:help': 'Randomize the initial order so users can\'t memorize positions'
    },
    showNumbers: {
      'ui:help': 'Display 1, 2, 3... next to items to help with ordering'
    },
    orientation: {
      'ui:help': 'Vertical works better for longer text items, horizontal for short labels'
    }
  }
};

export const OrderingEditor: FC<OrderingEditorProps> = ({ item, className }) => {
  return (
    <BasePropertyEditor
      item={item}
      schema={orderingSchema}
      uiSchema={orderingUiSchema}
      className={className}
    />
  );
};