import { FC } from 'react';
import { useCanvasStore } from '@/store/canvas-store';
import { ItemType } from '@/types/quiz';
import { cn } from '@/lib/utils';
import { Settings, AlertCircle, Info } from 'lucide-react';
import {
  ChoiceEditor,
  MultiChoiceEditor,
  TextEntryEditor,
  MediaEditor,
  TitleEditor,
  OrderingEditor,
  HotspotEditor
} from './property-editors';

interface PropertiesPanelProps {
  className?: string;
  style?: React.CSSProperties;
}

export const PropertiesPanel: FC<PropertiesPanelProps> = ({ 
  className, 
  style 
}) => {
  const { selectedItems, items } = useCanvasStore();
  
  const selectedItem = selectedItems.length === 1 
    ? items.find(item => item.id === selectedItems[0])
    : null;

  const renderPropertyEditor = () => {
    if (!selectedItem) return null;

    const commonProps = {
      item: selectedItem,
      className: 'mt-4'
    };

    switch (selectedItem.type) {
      case ItemType.TITLE:
        return <TitleEditor {...commonProps} />;
      case ItemType.MEDIA:
        return <MediaEditor {...commonProps} />;
      case ItemType.CHOICE:
        return <ChoiceEditor {...commonProps} />;
      case ItemType.MULTI_CHOICE:
        return <MultiChoiceEditor {...commonProps} />;
      case ItemType.TEXT_ENTRY:
        return <TextEntryEditor {...commonProps} />;
      case ItemType.ORDERING:
        return <OrderingEditor {...commonProps} />;
      case ItemType.HOTSPOT:
        return <HotspotEditor {...commonProps} />;
      default:
        return (
          <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-3 mt-4">
            <div className="flex items-center gap-2">
              <AlertCircle className="w-4 h-4 text-yellow-600" />
              <span className="text-sm font-medium text-yellow-800">
                Unsupported Item Type
              </span>
            </div>
            <p className="text-xs text-yellow-700 mt-1">
              No property editor available for type: {selectedItem.type}
            </p>
          </div>
        );
    }
  };

  return (
    <div 
      className={cn('bg-white overflow-y-auto', className)}
      style={style}
    >
      {/* Header */}
      <div className="p-4 border-b border-gray-200 bg-gray-50">
        <div className="flex items-center gap-2">
          <Settings className="w-5 h-5 text-gray-600" />
          <h2 className="text-lg font-semibold text-gray-900">Properties</h2>
        </div>
        {selectedItem && (
          <p className="text-xs text-gray-500 mt-1">
            Editing {selectedItem.type.replace('_', ' ')} • ID: {selectedItem.id.slice(-8)}
          </p>
        )}
      </div>

      <div className="p-4">
        {selectedItem ? (
          <div className="space-y-6">
            {/* Item Info Summary */}
            <div className="bg-blue-50 border border-blue-200 rounded-lg p-3">
              <div className="flex items-center gap-2 mb-2">
                <Info className="w-4 h-4 text-blue-600" />
                <span className="text-sm font-medium text-blue-800">Item Overview</span>
              </div>
              <div className="grid grid-cols-2 gap-3 text-xs">
                <div>
                  <span className="text-blue-600">Type:</span>
                  <span className="ml-1 font-medium capitalize text-blue-800">
                    {selectedItem.type.replace('_', ' ')}
                  </span>
                </div>
                <div>
                  <span className="text-blue-600">Position:</span>
                  <span className="ml-1 font-mono text-blue-800">
                    {Math.round(selectedItem.position.x)}, {Math.round(selectedItem.position.y)}
                  </span>
                </div>
                <div>
                  <span className="text-blue-600">Size:</span>
                  <span className="ml-1 font-mono text-blue-800">
                    {Math.round(selectedItem.size.width)} × {Math.round(selectedItem.size.height)}
                  </span>
                </div>
                <div>
                  <span className="text-blue-600">Z-Index:</span>
                  <span className="ml-1 font-mono text-blue-800">{selectedItem.zIndex}</span>
                </div>
              </div>
            </div>

            {/* Type-specific Property Editor */}
            <div className="border border-gray-200 rounded-lg">
              {renderPropertyEditor()}
            </div>
          </div>
        ) : selectedItems.length > 1 ? (
          <div className="text-center py-12">
            <Settings className="w-16 h-16 text-gray-300 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-600 mb-2">
              Multiple Items Selected
            </h3>
            <p className="text-sm text-gray-500 mb-4">
              {selectedItems.length} items are currently selected
            </p>
            <p className="text-xs text-gray-400">
              Select a single item to edit its properties
            </p>
          </div>
        ) : (
          <div className="text-center py-12">
            <Settings className="w-16 h-16 text-gray-300 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-600 mb-2">
              No Item Selected
            </h3>
            <p className="text-sm text-gray-500 mb-4">
              Click on an item in the canvas to edit its properties
            </p>
            <div className="bg-gray-50 rounded-lg p-4 text-left">
              <h4 className="text-xs font-medium text-gray-600 mb-2">Available Actions:</h4>
              <ul className="text-xs text-gray-500 space-y-1">
                <li>• Click an item to select and edit</li>
                <li>• Ctrl+Click for multi-selection</li>
                <li>• Drag items to reposition</li>
                <li>• Drag corners/edges to resize</li>
                <li>• Double-click for quick edit</li>
              </ul>
            </div>
          </div>
        )}
      </div>

      {/* Footer with validation status */}
      {selectedItem && (
        <div className="border-t border-gray-200 bg-gray-50 p-4">
          <div className="flex items-center justify-between text-xs">
            <span className="text-gray-500">
              Last updated: {new Date().toLocaleTimeString()}
            </span>
            <div className="flex items-center gap-1">
              <div className="w-2 h-2 bg-green-500 rounded-full"></div>
              <span className="text-green-600 font-medium">Valid</span>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};