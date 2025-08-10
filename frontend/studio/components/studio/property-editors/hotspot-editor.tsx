import { FC } from 'react';
import { BasePropertyEditor } from './base-property-editor';
import { CanvasItem } from '@/types/quiz';
import { RJSFSchema, UiSchema } from '@rjsf/utils';

interface HotspotEditorProps {
  item: CanvasItem;
  className?: string;
}

const hotspotSchema: RJSFSchema = {
  type: 'object',
  title: 'Hotspot Question Properties',
  properties: {
    title: {
      type: 'string',
      title: 'Question Text',
      description: 'Instructions for the hotspot interaction',
      minLength: 1,
      maxLength: 500
    },
    required: {
      type: 'boolean',
      title: 'Required Question',
      description: 'Users must click a hotspot to proceed',
      default: false
    },
    points: {
      type: 'number',
      title: 'Points',
      description: 'Points awarded for clicking correct hotspots',
      minimum: 0,
      maximum: 1000
    },
    explanation: {
      type: 'string',
      title: 'Explanation',
      description: 'Feedback explaining the correct hotspots (optional)',
      maxLength: 1000
    },
    content: {
      type: 'object',
      title: 'Hotspot Configuration',
      properties: {
        imageUrl: {
          type: 'string',
          title: 'Background Image URL',
          description: 'URL or path to the image with hotspot areas',
          format: 'uri'
        },
        altText: {
          type: 'string',
          title: 'Image Alt Text',
          description: 'Alternative text for accessibility',
          maxLength: 200
        },
        hotspots: {
          type: 'array',
          title: 'Hotspot Areas',
          description: 'Clickable areas on the image',
          minItems: 1,
          maxItems: 20,
          items: {
            type: 'object',
            title: 'Hotspot',
            properties: {
              id: {
                type: 'string',
                title: 'Hotspot ID',
                description: 'Unique identifier for this hotspot',
                default: () => `hotspot_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
              },
              shape: {
                type: 'string',
                title: 'Shape',
                description: 'Shape of the clickable area',
                enum: ['rectangle', 'circle', 'polygon'],
                enumNames: ['Rectangle', 'Circle', 'Polygon'],
                default: 'rectangle'
              },
              coords: {
                type: 'array',
                title: 'Coordinates',
                description: 'Position and size coordinates (format varies by shape)',
                items: {
                  type: 'number',
                  minimum: 0
                },
                minItems: 2,
                maxItems: 20
              },
              correct: {
                type: 'boolean',
                title: 'Correct Hotspot',
                description: 'Mark this hotspot as a correct answer',
                default: false
              },
              feedback: {
                type: 'string',
                title: 'Feedback Text',
                description: 'Message shown when this hotspot is clicked',
                maxLength: 200
              }
            },
            required: ['id', 'shape', 'coords']
          },
          default: [
            {
              id: 'hotspot_1',
              shape: 'rectangle',
              coords: [100, 100, 200, 150],
              correct: true,
              feedback: 'Correct! You found the right area.'
            },
            {
              id: 'hotspot_2',
              shape: 'circle',
              coords: [300, 200, 50],
              correct: false,
              feedback: 'Not quite right. Try another area.'
            }
          ]
        },
        multipleSelect: {
          type: 'boolean',
          title: 'Allow Multiple Selections',
          description: 'Users can click multiple hotspots before submitting',
          default: false
        },
        showFeedback: {
          type: 'boolean',
          title: 'Show Immediate Feedback',
          description: 'Display feedback immediately when hotspots are clicked',
          default: true
        },
        highlightOnHover: {
          type: 'boolean',
          title: 'Highlight on Hover',
          description: 'Show visual indication when hovering over hotspots',
          default: true
        }
      },
      required: ['imageUrl', 'hotspots']
    }
  },
  required: ['title', 'required', 'content']
};

const hotspotUiSchema: UiSchema = {
  title: {
    'ui:widget': 'textarea',
    'ui:options': {
      rows: 2
    },
    'ui:placeholder': 'e.g., Click on the correct area in the image...'
  },
  explanation: {
    'ui:widget': 'textarea',
    'ui:options': {
      rows: 3
    },
    'ui:placeholder': 'Explain why certain areas are correct...'
  },
  content: {
    'ui:title': 'Hotspot Settings',
    'ui:order': ['imageUrl', 'altText', 'hotspots', 'multipleSelect', 'showFeedback', 'highlightOnHover'],
    imageUrl: {
      'ui:placeholder': 'https://example.com/image.jpg or upload://diagram.png'
    },
    altText: {
      'ui:placeholder': 'Describe what is shown in the image...',
      'ui:help': 'Important for accessibility'
    },
    hotspots: {
      'ui:options': {
        addable: true,
        removable: true,
        orderable: true
      },
      items: {
        'ui:order': ['shape', 'coords', 'correct', 'feedback', 'id'],
        id: {
          'ui:widget': 'hidden'
        },
        shape: {
          'ui:help': 'Rectangle: [x, y, width, height] | Circle: [centerX, centerY, radius] | Polygon: [x1, y1, x2, y2, ...]'
        },
        coords: {
          'ui:help': 'Coordinates depend on shape type. Use image editor to find exact positions.',
          items: {
            'ui:placeholder': '0'
          }
        },
        correct: {
          'ui:help': 'Mark correct hotspots that award points'
        },
        feedback: {
          'ui:placeholder': 'Message shown when clicked...',
          'ui:widget': 'textarea',
          'ui:options': {
            rows: 2
          }
        }
      }
    },
    multipleSelect: {
      'ui:help': 'Allow users to select multiple areas before submitting their answer'
    },
    showFeedback: {
      'ui:help': 'Show feedback text immediately, or wait until all selections are made'
    },
    highlightOnHover: {
      'ui:help': 'Help users discover clickable areas by highlighting them on hover'
    }
  }
};

export const HotspotEditor: FC<HotspotEditorProps> = ({ item, className }) => {
  return (
    <BasePropertyEditor
      item={item}
      schema={hotspotSchema}
      uiSchema={hotspotUiSchema}
      className={className}
    />
  );
};