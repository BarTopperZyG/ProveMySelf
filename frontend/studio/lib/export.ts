import { CanvasItem, ItemType, QuizBundle } from '@/types/quiz';
import { AssetManager } from './asset-manager';

export interface AdaptiveCardAction {
  type: string;
  title: string;
  id: string;
  style?: string;
}

export interface AdaptiveCardElement {
  type: string;
  id?: string;
  text?: string;
  url?: string;
  altText?: string;
  size?: string;
  weight?: string;
  color?: string;
  horizontalAlignment?: string;
  spacing?: string;
  choices?: Array<{
    title: string;
    value: string;
  }>;
  isMultiSelect?: boolean;
  style?: string;
  placeholder?: string;
  maxLength?: number;
  actions?: AdaptiveCardAction[];
  items?: AdaptiveCardElement[];
  columns?: Array<{
    type: string;
    width: string;
    items: AdaptiveCardElement[];
  }>;
}

export interface AdaptiveCard {
  $schema: string;
  type: string;
  version: string;
  body: AdaptiveCardElement[];
  actions?: AdaptiveCardAction[];
}

export interface QTIItem {
  '@xmlns': string;
  '@identifier': string;
  '@title': string;
  '@adaptive': boolean;
  '@timeDependent': boolean;
  responseDeclaration?: Array<{
    '@identifier': string;
    '@cardinality': string;
    '@baseType': string;
    correctResponse?: {
      value: string | string[];
    };
    mapping?: {
      mapEntry: Array<{
        '@mapKey': string;
        '@mappedValue': number;
      }>;
    };
  }>;
  outcomeDeclaration?: Array<{
    '@identifier': string;
    '@cardinality': string;
    '@baseType': string;
    defaultValue?: {
      value: number;
    };
  }>;
  itemBody: {
    div: string | object;
  };
  responseProcessing?: {
    setOutcomeValue?: Array<{
      '@identifier': string;
      sum?: {
        variable: Array<{
          '@identifier': string;
        }>;
      };
    }>;
  };
}

export interface QTIAssessmentItem {
  assessmentItem: QTIItem;
}

/**
 * Converts canvas items to Adaptive Cards format
 */
export function generateAdaptiveCards(items: CanvasItem[]): AdaptiveCard[] {
  const cards: AdaptiveCard[] = [];
  
  // Sort items by position and group by page/screen
  const sortedItems = items
    .filter(item => item.type !== ItemType.TITLE || items.filter(i => i.type === ItemType.TITLE).length === 1)
    .sort((a, b) => a.position.y - b.position.y || a.position.x - b.position.x);

  // Group items into logical cards
  let currentCard: AdaptiveCard = {
    $schema: 'http://adaptivecards.io/schemas/adaptive-card.json',
    type: 'AdaptiveCard',
    version: '1.5',
    body: []
  };

  for (const item of sortedItems) {
    const cardElement = convertItemToAdaptiveCardElement(item);
    if (cardElement) {
      currentCard.body.push(cardElement);
    }

    // Create new card for each question (not titles/media)
    if (isQuestionType(item.type) && currentCard.body.length > 0) {
      cards.push(currentCard);
      currentCard = {
        $schema: 'http://adaptivecards.io/schemas/adaptive-card.json',
        type: 'AdaptiveCard',
        version: '1.5',
        body: []
      };
    }
  }

  // Add final card if it has content
  if (currentCard.body.length > 0) {
    cards.push(currentCard);
  }

  return cards.length > 0 ? cards : [currentCard];
}

/**
 * Converts canvas items to QTI 3.0 format
 */
export function generateQTIItems(items: CanvasItem[]): QTIAssessmentItem[] {
  const qtiItems: QTIAssessmentItem[] = [];
  let itemCounter = 1;

  const questionItems = items.filter(item => isQuestionType(item.type));

  for (const item of questionItems) {
    const qtiItem = convertItemToQTIItem(item, itemCounter);
    if (qtiItem) {
      qtiItems.push(qtiItem);
      itemCounter++;
    }
  }

  return qtiItems;
}

/**
 * Converts a canvas item to an Adaptive Card element
 */
function convertItemToAdaptiveCardElement(item: CanvasItem): AdaptiveCardElement | null {
  switch (item.type) {
    case ItemType.TITLE:
      return {
        type: 'TextBlock',
        text: item.title,
        size: getAdaptiveCardSize(item.content?.level || 2),
        weight: item.content?.emphasis === 'bold' ? 'Bolder' : 'Default',
        color: getAdaptiveCardColor(item.content?.color || 'default'),
        horizontalAlignment: item.content?.alignment || 'Left'
      };

    case ItemType.MEDIA:
      if (item.content?.mediaType === 'image') {
        return {
          type: 'Image',
          url: item.content?.url || '',
          altText: item.content?.altText || item.title,
          size: 'Auto'
        };
      } else if (item.content?.mediaType === 'video') {
        return {
          type: 'Media',
          sources: [{
            mimeType: 'video/mp4',
            url: item.content?.url || ''
          }],
          altText: item.content?.altText || item.title
        };
      }
      break;

    case ItemType.CHOICE:
      return {
        type: 'Input.ChoiceSet',
        id: item.id,
        choices: (item.content?.choices || []).map((choice: any, index: number) => ({
          title: choice.text,
          value: choice.id || `choice_${index}`
        })),
        style: 'expanded',
        isMultiSelect: false
      };

    case ItemType.MULTI_CHOICE:
      return {
        type: 'Input.ChoiceSet',
        id: item.id,
        choices: (item.content?.choices || []).map((choice: any, index: number) => ({
          title: choice.text,
          value: choice.id || `choice_${index}`
        })),
        style: 'expanded',
        isMultiSelect: true
      };

    case ItemType.TEXT_ENTRY:
      return {
        type: 'Input.Text',
        id: item.id,
        placeholder: item.content?.placeholder || 'Enter your answer...',
        maxLength: item.content?.maxLength || 1000,
        isMultiline: (item.content?.maxLength || 100) > 100
      };

    case ItemType.ORDERING:
      // Adaptive Cards doesn't have native ordering, so we'll use a choice set with instructions
      return {
        type: 'Container',
        items: [
          {
            type: 'TextBlock',
            text: item.title + ' (Select items in correct order)',
            weight: 'Bolder'
          },
          {
            type: 'Input.ChoiceSet',
            id: item.id,
            choices: (item.content?.items || [])
              .sort(() => Math.random() - 0.5) // Shuffle for ordering
              .map((orderItem: any, index: number) => ({
                title: `${index + 1}. ${orderItem.text}`,
                value: orderItem.id
              })),
            style: 'expanded',
            isMultiSelect: true
          }
        ]
      };

    case ItemType.HOTSPOT:
      // Adaptive Cards doesn't support hotspots natively, so we'll create a choice set
      return {
        type: 'Container',
        items: [
          {
            type: 'TextBlock',
            text: item.title,
            weight: 'Bolder'
          },
          {
            type: 'Image',
            url: item.content?.imageUrl || '',
            altText: item.content?.altText || 'Hotspot image'
          },
          {
            type: 'Input.ChoiceSet',
            id: item.id,
            choices: (item.content?.hotspots || []).map((hotspot: any, index: number) => ({
              title: hotspot.feedback || `Area ${index + 1}`,
              value: hotspot.id
            })),
            style: 'expanded',
            isMultiSelect: item.content?.multipleSelect || false
          }
        ]
      };

    default:
      return null;
  }

  return null;
}

/**
 * Converts a canvas item to a QTI assessment item
 */
function convertItemToQTIItem(item: CanvasItem, itemNumber: number): QTIAssessmentItem | null {
  const identifier = `item_${itemNumber}`;
  
  switch (item.type) {
    case ItemType.CHOICE:
    case ItemType.MULTI_CHOICE:
      return {
        assessmentItem: {
          '@xmlns': 'http://www.imsglobal.org/xsd/qti/qtiv3p0/imsqtiasi_v3p0',
          '@identifier': identifier,
          '@title': item.title,
          '@adaptive': false,
          '@timeDependent': false,
          responseDeclaration: [{
            '@identifier': 'RESPONSE',
            '@cardinality': item.type === ItemType.MULTI_CHOICE ? 'multiple' : 'single',
            '@baseType': 'identifier',
            correctResponse: {
              value: getCorrectResponses(item)
            }
          }],
          outcomeDeclaration: [{
            '@identifier': 'SCORE',
            '@cardinality': 'single',
            '@baseType': 'float',
            defaultValue: {
              value: 0
            }
          }],
          itemBody: {
            div: generateQTIItemBody(item)
          },
          responseProcessing: {
            setOutcomeValue: [{
              '@identifier': 'SCORE',
              sum: {
                variable: [{
                  '@identifier': 'RESPONSE'
                }]
              }
            }]
          }
        }
      };

    case ItemType.TEXT_ENTRY:
      return {
        assessmentItem: {
          '@xmlns': 'http://www.imsglobal.org/xsd/qti/qtiv3p0/imsqtiasi_v3p0',
          '@identifier': identifier,
          '@title': item.title,
          '@adaptive': false,
          '@timeDependent': false,
          responseDeclaration: [{
            '@identifier': 'RESPONSE',
            '@cardinality': 'single',
            '@baseType': 'string'
          }],
          outcomeDeclaration: [{
            '@identifier': 'SCORE',
            '@cardinality': 'single',
            '@baseType': 'float',
            defaultValue: {
              value: 0
            }
          }],
          itemBody: {
            div: generateQTIItemBody(item)
          }
        }
      };

    default:
      return null;
  }
}

/**
 * Helper functions
 */
function isQuestionType(type: ItemType): boolean {
  return [
    ItemType.CHOICE,
    ItemType.MULTI_CHOICE,
    ItemType.TEXT_ENTRY,
    ItemType.ORDERING,
    ItemType.HOTSPOT
  ].includes(type);
}

function getAdaptiveCardSize(level: number): string {
  const sizeMap: { [key: number]: string } = {
    1: 'ExtraLarge',
    2: 'Large',
    3: 'Medium',
    4: 'Default',
    5: 'Small',
    6: 'Small'
  };
  return sizeMap[level] || 'Default';
}

function getAdaptiveCardColor(color: string): string {
  const colorMap: { [key: string]: string } = {
    'default': 'Default',
    'primary': 'Accent',
    'secondary': 'Good',
    'success': 'Good',
    'warning': 'Warning',
    'error': 'Attention'
  };
  return colorMap[color] || 'Default';
}

function getCorrectResponses(item: CanvasItem): string | string[] {
  if (item.type === ItemType.CHOICE) {
    const correctChoice = (item.content?.choices || []).find((choice: any) => choice.correct);
    return correctChoice ? correctChoice.id : '';
  } else if (item.type === ItemType.MULTI_CHOICE) {
    return (item.content?.choices || [])
      .filter((choice: any) => choice.correct)
      .map((choice: any) => choice.id);
  }
  return '';
}

function generateQTIItemBody(item: CanvasItem): string {
  let body = `<p>${item.title}</p>`;
  
  switch (item.type) {
    case ItemType.CHOICE:
      body += '<choiceInteraction responseIdentifier="RESPONSE" shuffle="true" maxChoices="1">';
      (item.content?.choices || []).forEach((choice: any) => {
        body += `<simpleChoice identifier="${choice.id}">${choice.text}</simpleChoice>`;
      });
      body += '</choiceInteraction>';
      break;

    case ItemType.MULTI_CHOICE:
      body += '<choiceInteraction responseIdentifier="RESPONSE" shuffle="true" maxChoices="0">';
      (item.content?.choices || []).forEach((choice: any) => {
        body += `<simpleChoice identifier="${choice.id}">${choice.text}</simpleChoice>`;
      });
      body += '</choiceInteraction>';
      break;

    case ItemType.TEXT_ENTRY:
      body += `<extendedTextInteraction responseIdentifier="RESPONSE" expectedLength="${item.content?.maxLength || 100}"/>`;
      break;
  }

  return body;
}

/**
 * Creates a complete quiz bundle with assets
 */
export async function createQuizBundle(items: CanvasItem[], metadata?: any): Promise<QuizBundle> {
  const assetManager = new AssetManager();
  
  // Process assets
  const assetBundle = await assetManager.createAssetBundle(items);
  
  // Localize asset references in items
  const localizedItems = assetManager.localizeAssetReferences(items, assetBundle.assets);
  
  return {
    metadata: {
      id: metadata?.id || `quiz_${Date.now()}`,
      title: metadata?.title || 'Untitled Quiz',
      description: metadata?.description || 'Generated quiz from Canvas Studio',
      version: '1.0.0',
      created: new Date().toISOString(),
      author: metadata?.author || 'Canvas Studio',
      tags: metadata?.tags || [],
      estimatedTime: metadata?.estimatedTime || calculateEstimatedTime(items)
    },
    ui: generateAdaptiveCards(localizedItems),
    quiz: generateQTIItems(localizedItems),
    assets: assetBundle.assets.map(asset => ({
      id: asset.id,
      name: asset.name,
      type: asset.type,
      mimeType: asset.mimeType,
      size: asset.size,
      url: asset.url,
      metadata: asset.metadata
    })),
    settings: {
      allowBackNavigation: metadata?.settings?.allowBackNavigation ?? true,
      showProgressBar: metadata?.settings?.showProgressBar ?? true,
      randomizeQuestions: metadata?.settings?.randomizeQuestions ?? false,
      timeLimit: metadata?.settings?.timeLimit,
      maxAttempts: metadata?.settings?.maxAttempts,
      passingScore: metadata?.settings?.passingScore
    }
  };
}

/**
 * Creates a complete downloadable bundle
 */
export async function createDownloadableBundle(
  items: CanvasItem[], 
  metadata?: any
): Promise<{ bundle: QuizBundle; blob: Blob }> {
  const assetManager = new AssetManager();
  const bundle = await createQuizBundle(items, metadata);
  const blob = await assetManager.createDownloadableBundle(
    items,
    bundle.ui,
    bundle.quiz,
    bundle.metadata
  );
  
  return { bundle, blob };
}

/**
 * Estimates completion time based on question types and content
 */
function calculateEstimatedTime(items: CanvasItem[]): number {
  const timePerType: { [key in ItemType]: number } = {
    [ItemType.TITLE]: 5,
    [ItemType.MEDIA]: 30,
    [ItemType.CHOICE]: 15,
    [ItemType.MULTI_CHOICE]: 20,
    [ItemType.TEXT_ENTRY]: 60,
    [ItemType.ORDERING]: 30,
    [ItemType.HOTSPOT]: 25
  };

  let totalTime = 0;
  for (const item of items) {
    totalTime += timePerType[item.type] || 15;
  }

  return Math.ceil(totalTime / 60); // Return in minutes
}

/**
 * Validates quiz bundle structure
 */
export function validateQuizBundle(bundle: QuizBundle): { isValid: boolean; errors: string[] } {
  const errors: string[] = [];

  if (!bundle.metadata?.id) {
    errors.push('Bundle must have a metadata ID');
  }

  if (!bundle.metadata?.title) {
    errors.push('Bundle must have a title');
  }

  if (!bundle.ui || bundle.ui.length === 0) {
    errors.push('Bundle must have UI cards');
  }

  if (!bundle.quiz || bundle.quiz.length === 0) {
    errors.push('Bundle must have quiz items');
  }

  // Validate UI cards structure
  bundle.ui?.forEach((card, index) => {
    if (!card.$schema) {
      errors.push(`UI card ${index} missing schema`);
    }
    if (!card.body || card.body.length === 0) {
      errors.push(`UI card ${index} has no body elements`);
    }
  });

  // Validate QTI items structure
  bundle.quiz?.forEach((item, index) => {
    if (!item.assessmentItem?.['@identifier']) {
      errors.push(`QTI item ${index} missing identifier`);
    }
    if (!item.assessmentItem?.itemBody) {
      errors.push(`QTI item ${index} missing item body`);
    }
  });

  return {
    isValid: errors.length === 0,
    errors
  };
}