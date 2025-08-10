import { FC, useState } from 'react';
import { useCanvasStore } from '@/store/canvas-store';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { 
  Play, 
  Pause, 
  RotateCcw, 
  ChevronLeft, 
  ChevronRight, 
  Monitor, 
  Tablet, 
  Smartphone,
  Eye
} from 'lucide-react';

interface PreviewPanelProps {
  className?: string;
}

type PreviewDevice = 'desktop' | 'tablet' | 'mobile';

const DEVICE_SIZES = {
  desktop: { width: '100%', height: '100%' },
  tablet: { width: '768px', height: '1024px' },
  mobile: { width: '375px', height: '667px' }
};

const DEVICE_ICONS = {
  desktop: Monitor,
  tablet: Tablet,
  mobile: Smartphone
};

export const PreviewPanel: FC<PreviewPanelProps> = ({ className }) => {
  const { items, canvasSize } = useCanvasStore();
  const [currentDevice, setCurrentDevice] = useState<PreviewDevice>('desktop');
  const [currentItemIndex, setCurrentItemIndex] = useState(0);
  const [isPlaying, setIsPlaying] = useState(false);

  const sortedItems = [...items].sort((a, b) => a.position.y - b.position.y || a.position.x - b.position.x);

  const handlePrevious = () => {
    setCurrentItemIndex(Math.max(0, currentItemIndex - 1));
  };

  const handleNext = () => {
    setCurrentItemIndex(Math.min(sortedItems.length - 1, currentItemIndex + 1));
  };

  const handleReset = () => {
    setCurrentItemIndex(0);
    setIsPlaying(false);
  };

  const togglePlayback = () => {
    setIsPlaying(!isPlaying);
  };

  const currentItem = sortedItems[currentItemIndex];

  return (
    <div className={cn('flex flex-col h-full bg-gray-100', className)}>
      {/* Preview Controls */}
      <div className="flex items-center justify-between px-4 py-3 bg-white border-b border-gray-200">
        {/* Device Selection */}
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium text-gray-600">Device:</span>
          <div className="flex items-center bg-gray-100 rounded-lg p-1">
            {(Object.keys(DEVICE_SIZES) as PreviewDevice[]).map((device) => {
              const Icon = DEVICE_ICONS[device];
              return (
                <Button
                  key={device}
                  variant={currentDevice === device ? "default" : "ghost"}
                  size="sm"
                  onClick={() => setCurrentDevice(device)}
                  className="gap-2 capitalize"
                >
                  <Icon className="w-4 h-4" />
                  {device}
                </Button>
              );
            })}
          </div>
        </div>

        {/* Navigation Controls */}
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={handlePrevious}
            disabled={currentItemIndex === 0}
          >
            <ChevronLeft className="w-4 h-4" />
            Previous
          </Button>
          
          <span className="text-sm text-gray-600 px-2">
            {sortedItems.length > 0 ? `${currentItemIndex + 1} of ${sortedItems.length}` : '0 of 0'}
          </span>
          
          <Button
            variant="outline"
            size="sm"
            onClick={handleNext}
            disabled={currentItemIndex >= sortedItems.length - 1}
          >
            Next
            <ChevronRight className="w-4 h-4" />
          </Button>
        </div>

        {/* Playback Controls */}
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={togglePlayback}
            disabled={sortedItems.length === 0}
            className="gap-2"
          >
            {isPlaying ? <Pause className="w-4 h-4" /> : <Play className="w-4 h-4" />}
            {isPlaying ? 'Pause' : 'Start'}
          </Button>
          
          <Button
            variant="ghost"
            size="sm"
            onClick={handleReset}
            title="Reset to beginning"
          >
            <RotateCcw className="w-4 h-4" />
          </Button>
        </div>
      </div>

      {/* Preview Content */}
      <div className="flex-1 flex items-center justify-center p-8 overflow-auto">
        {sortedItems.length > 0 ? (
          <div 
            className="bg-white rounded-lg shadow-lg overflow-hidden transition-all duration-300"
            style={{
              width: DEVICE_SIZES[currentDevice].width,
              height: DEVICE_SIZES[currentDevice].height,
              maxWidth: '100%',
              maxHeight: '100%'
            }}
          >
            {/* Preview Header */}
            <div className="bg-blue-600 text-white p-4 text-center">
              <h1 className="text-xl font-bold">Quiz Preview</h1>
              <p className="text-sm opacity-90 mt-1">
                {currentItem ? `Question ${currentItemIndex + 1}` : 'No questions'}
              </p>
            </div>

            {/* Question Content */}
            <div className="p-6 min-h-96">
              {currentItem ? (
                <div className="space-y-6">
                  {/* Question Title */}
                  <div>
                    <h2 className="text-lg font-semibold text-gray-900 mb-2">
                      {currentItem.title}
                    </h2>
                    {currentItem.required && (
                      <span className="inline-block bg-red-100 text-red-800 text-xs px-2 py-1 rounded-full">
                        Required
                      </span>
                    )}
                    {currentItem.points && (
                      <span className="inline-block bg-blue-100 text-blue-800 text-xs px-2 py-1 rounded-full ml-2">
                        {currentItem.points} points
                      </span>
                    )}
                  </div>

                  {/* Question Content Based on Type */}
                  <div className="bg-gray-50 p-4 rounded-lg">
                    {currentItem.type === 'title' && (
                      <div className="text-center py-8">
                        <h3 className="text-2xl font-bold text-gray-800">
                          {currentItem.title}
                        </h3>
                        <p className="text-gray-600 mt-2">Title Block</p>
                      </div>
                    )}

                    {currentItem.type === 'choice' && (
                      <div className="space-y-3">
                        <p className="text-gray-700 mb-4">Select one option:</p>
                        <div className="space-y-2">
                          {['Option A', 'Option B', 'Option C', 'Option D'].map((option, index) => (
                            <label key={index} className="flex items-center p-3 border rounded-lg hover:bg-white cursor-pointer">
                              <input
                                type="radio"
                                name="choice"
                                className="mr-3 text-blue-600"
                              />
                              <span>{option}</span>
                            </label>
                          ))}
                        </div>
                      </div>
                    )}

                    {currentItem.type === 'multi_choice' && (
                      <div className="space-y-3">
                        <p className="text-gray-700 mb-4">Select all that apply:</p>
                        <div className="space-y-2">
                          {['Option A', 'Option B', 'Option C', 'Option D'].map((option, index) => (
                            <label key={index} className="flex items-center p-3 border rounded-lg hover:bg-white cursor-pointer">
                              <input
                                type="checkbox"
                                className="mr-3 text-blue-600 rounded"
                              />
                              <span>{option}</span>
                            </label>
                          ))}
                        </div>
                      </div>
                    )}

                    {currentItem.type === 'text_entry' && (
                      <div className="space-y-3">
                        <p className="text-gray-700 mb-4">Enter your answer:</p>
                        <textarea
                          className="w-full p-3 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                          rows={4}
                          placeholder="Type your answer here..."
                        />
                      </div>
                    )}

                    {currentItem.type === 'media' && (
                      <div className="text-center py-8">
                        <div className="w-full h-48 bg-gray-200 rounded-lg flex items-center justify-center">
                          <div className="text-gray-500">
                            <Eye className="w-8 h-8 mx-auto mb-2" />
                            <p>Media Content</p>
                            <p className="text-sm">Image, Video, or Audio</p>
                          </div>
                        </div>
                      </div>
                    )}

                    {currentItem.type === 'ordering' && (
                      <div className="space-y-3">
                        <p className="text-gray-700 mb-4">Drag items to arrange in correct order:</p>
                        <div className="space-y-2">
                          {['First item', 'Second item', 'Third item', 'Fourth item'].map((item, index) => (
                            <div
                              key={index}
                              className="flex items-center p-3 border rounded-lg bg-white cursor-move"
                            >
                              <div className="flex-1">{item}</div>
                              <div className="text-gray-400">≡</div>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}

                    {currentItem.type === 'hotspot' && (
                      <div className="text-center py-8">
                        <div className="w-full h-64 bg-gray-200 rounded-lg flex items-center justify-center relative">
                          <div className="text-gray-500">
                            <p>Hotspot Image</p>
                            <p className="text-sm">Click areas on image</p>
                          </div>
                          {/* Example hotspot areas */}
                          <div className="absolute top-4 left-8 w-8 h-8 border-2 border-red-500 rounded-full bg-red-100 opacity-50"></div>
                          <div className="absolute bottom-8 right-12 w-12 h-8 border-2 border-blue-500 bg-blue-100 opacity-50"></div>
                        </div>
                      </div>
                    )}
                  </div>

                  {/* Explanation */}
                  {currentItem.explanation && (
                    <div className="bg-blue-50 p-4 rounded-lg">
                      <h4 className="font-medium text-blue-900 mb-2">Explanation</h4>
                      <p className="text-blue-800 text-sm">{currentItem.explanation}</p>
                    </div>
                  )}
                </div>
              ) : (
                <div className="text-center py-8 text-gray-500">
                  <Eye className="w-12 h-12 mx-auto mb-3 opacity-50" />
                  <p>No questions to preview</p>
                  <p className="text-sm mt-1">Add some question blocks to see them here</p>
                </div>
              )}
            </div>

            {/* Preview Footer */}
            {sortedItems.length > 0 && (
              <div className="bg-gray-50 p-4 border-t flex justify-between items-center">
                <div className="text-sm text-gray-600">
                  Progress: {Math.round(((currentItemIndex + 1) / sortedItems.length) * 100)}%
                </div>
                <div className="flex gap-2">
                  <Button variant="outline" size="sm" disabled={currentItemIndex === 0}>
                    Back
                  </Button>
                  <Button size="sm" disabled={currentItemIndex >= sortedItems.length - 1}>
                    {currentItemIndex >= sortedItems.length - 1 ? 'Submit' : 'Continue'}
                  </Button>
                </div>
              </div>
            )}
          </div>
        ) : (
          <div className="text-center py-12">
            <Eye className="w-16 h-16 text-gray-300 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-600 mb-2">No questions to preview</h3>
            <p className="text-gray-500">
              Switch to Edit mode and add some question blocks to see them here
            </p>
          </div>
        )}
      </div>
    </div>
  );
};