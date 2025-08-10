import { FC, useState } from 'react';
import { useCanvasStore } from '@/store/canvas-store';
import { createQuizBundle, validateQuizBundle, createDownloadableBundle } from '@/lib/export';
import { cn } from '@/lib/utils';
import {
  Download,
  FileJson,
  Package,
  Settings,
  AlertTriangle,
  CheckCircle,
  X,
  Info
} from 'lucide-react';

interface ExportDialogProps {
  isOpen: boolean;
  onClose: () => void;
}

interface ExportSettings {
  title: string;
  description: string;
  author: string;
  tags: string[];
  estimatedTime?: number;
  allowBackNavigation: boolean;
  showProgressBar: boolean;
  randomizeQuestions: boolean;
  timeLimit?: number;
  maxAttempts?: number;
  passingScore?: number;
}

export const ExportDialog: FC<ExportDialogProps> = ({ isOpen, onClose }) => {
  const { items } = useCanvasStore();
  const [exportSettings, setExportSettings] = useState<ExportSettings>({
    title: 'My Quiz',
    description: 'A quiz created with Canvas Studio',
    author: 'Quiz Author',
    tags: [],
    allowBackNavigation: true,
    showProgressBar: true,
    randomizeQuestions: false
  });
  const [newTag, setNewTag] = useState('');
  const [isExporting, setIsExporting] = useState(false);
  const [exportResult, setExportResult] = useState<{
    success: boolean;
    message: string;
    errors?: string[];
  } | null>(null);

  if (!isOpen) return null;

  const handleExport = async (format: 'json' | 'bundle') => {
    setIsExporting(true);
    setExportResult(null);

    try {
      const metadata = {
        id: `quiz_${Date.now()}`,
        title: exportSettings.title,
        description: exportSettings.description,
        author: exportSettings.author,
        tags: exportSettings.tags,
        estimatedTime: exportSettings.estimatedTime,
        settings: {
          allowBackNavigation: exportSettings.allowBackNavigation,
          showProgressBar: exportSettings.showProgressBar,
          randomizeQuestions: exportSettings.randomizeQuestions,
          timeLimit: exportSettings.timeLimit,
          maxAttempts: exportSettings.maxAttempts,
          passingScore: exportSettings.passingScore
        }
      };

      if (format === 'json') {
        // Export as separate JSON files with assets processed
        const bundle = await createQuizBundle(items, metadata);
        const validation = validateQuizBundle(bundle);
        
        if (!validation.isValid) {
          setExportResult({
            success: false,
            message: 'Export validation failed',
            errors: validation.errors
          });
          return;
        }

        // Create and download separate files
        const uiJson = JSON.stringify(bundle.ui, null, 2);
        const quizJson = JSON.stringify(bundle.quiz, null, 2);
        const metadataJson = JSON.stringify(bundle.metadata, null, 2);
        const assetsJson = JSON.stringify(bundle.assets, null, 2);

        downloadFile(`${bundle.metadata.id}_ui.json`, uiJson, 'application/json');
        downloadFile(`${bundle.metadata.id}_quiz.json`, quizJson, 'application/json');
        downloadFile(`${bundle.metadata.id}_metadata.json`, metadataJson, 'application/json');
        
        if (bundle.assets && bundle.assets.length > 0) {
          downloadFile(`${bundle.metadata.id}_assets.json`, assetsJson, 'application/json');
        }

        setExportResult({
          success: true,
          message: `Quiz exported as JSON files successfully! ${bundle.assets?.length || 0} assets processed.`
        });
      } else {
        // Export as complete downloadable bundle
        const { bundle, blob } = await createDownloadableBundle(items, metadata);
        const validation = validateQuizBundle(bundle);
        
        if (!validation.isValid) {
          setExportResult({
            success: false,
            message: 'Export validation failed',
            errors: validation.errors
          });
          return;
        }

        // Download the complete bundle
        downloadBlob(`${bundle.metadata.id}_complete_bundle.json`, blob);

        setExportResult({
          success: true,
          message: `Complete quiz bundle exported successfully! Includes ${bundle.assets?.length || 0} assets (${formatFileSize(blob.size)}).`
        });
      }
    } catch (error) {
      setExportResult({
        success: false,
        message: `Export failed: ${error instanceof Error ? error.message : 'Unknown error'}`
      });
    } finally {
      setIsExporting(false);
    }
  };

  const downloadFile = (filename: string, content: string, mimeType: string) => {
    const blob = new Blob([content], { type: mimeType });
    downloadBlob(filename, blob);
  };

  const downloadBlob = (filename: string, blob: Blob) => {
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const addTag = () => {
    if (newTag.trim() && !exportSettings.tags.includes(newTag.trim())) {
      setExportSettings({
        ...exportSettings,
        tags: [...exportSettings.tags, newTag.trim()]
      });
      setNewTag('');
    }
  };

  const removeTag = (tagToRemove: string) => {
    setExportSettings({
      ...exportSettings,
      tags: exportSettings.tags.filter(tag => tag !== tagToRemove)
    });
  };

  const questionCount = items.filter(item => 
    ['CHOICE', 'MULTI_CHOICE', 'TEXT_ENTRY', 'ORDERING', 'HOTSPOT'].includes(item.type)
  ).length;

  return (
    <div className="fixed inset-0 z-50 bg-black bg-opacity-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-lg max-w-2xl w-full max-h-[90vh] overflow-hidden">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-gray-200">
          <div className="flex items-center gap-3">
            <Package className="w-6 h-6 text-blue-600" />
            <h2 className="text-xl font-semibold text-gray-900">Export Quiz</h2>
          </div>
          <button
            onClick={onClose}
            className="p-2 hover:bg-gray-100 rounded-lg"
          >
            <X className="w-5 h-5 text-gray-500" />
          </button>
        </div>

        <div className="overflow-y-auto max-h-[calc(90vh-140px)]">
          {/* Quiz Summary */}
          <div className="p-6 border-b border-gray-200">
            <div className="flex items-center gap-2 mb-4">
              <Info className="w-5 h-5 text-blue-600" />
              <h3 className="text-lg font-medium text-gray-900">Quiz Summary</h3>
            </div>
            <div className="grid grid-cols-3 gap-4 text-sm">
              <div className="bg-gray-50 p-3 rounded-lg">
                <div className="text-gray-600">Total Items</div>
                <div className="text-2xl font-bold text-gray-900">{items.length}</div>
              </div>
              <div className="bg-blue-50 p-3 rounded-lg">
                <div className="text-blue-600">Questions</div>
                <div className="text-2xl font-bold text-blue-900">{questionCount}</div>
              </div>
              <div className="bg-green-50 p-3 rounded-lg">
                <div className="text-green-600">Est. Time</div>
                <div className="text-2xl font-bold text-green-900">
                  {Math.ceil((questionCount * 20 + (items.length - questionCount) * 5) / 60)}m
                </div>
              </div>
            </div>
          </div>

          {/* Export Settings */}
          <div className="p-6">
            <div className="flex items-center gap-2 mb-4">
              <Settings className="w-5 h-5 text-gray-600" />
              <h3 className="text-lg font-medium text-gray-900">Export Settings</h3>
            </div>

            <div className="space-y-4">
              {/* Basic Information */}
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Quiz Title
                  </label>
                  <input
                    type="text"
                    value={exportSettings.title}
                    onChange={(e) => setExportSettings({
                      ...exportSettings,
                      title: e.target.value
                    })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Author
                  </label>
                  <input
                    type="text"
                    value={exportSettings.author}
                    onChange={(e) => setExportSettings({
                      ...exportSettings,
                      author: e.target.value
                    })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Description
                </label>
                <textarea
                  value={exportSettings.description}
                  onChange={(e) => setExportSettings({
                    ...exportSettings,
                    description: e.target.value
                  })}
                  rows={3}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>

              {/* Tags */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Tags
                </label>
                <div className="flex gap-2 mb-2">
                  <input
                    type="text"
                    value={newTag}
                    onChange={(e) => setNewTag(e.target.value)}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter') {
                        e.preventDefault();
                        addTag();
                      }
                    }}
                    placeholder="Add a tag..."
                    className="flex-1 px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                  <button
                    onClick={addTag}
                    className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
                  >
                    Add
                  </button>
                </div>
                <div className="flex flex-wrap gap-2">
                  {exportSettings.tags.map((tag, index) => (
                    <span
                      key={index}
                      className="inline-flex items-center gap-1 px-2 py-1 bg-gray-100 text-gray-700 rounded text-sm"
                    >
                      {tag}
                      <button
                        onClick={() => removeTag(tag)}
                        className="text-gray-500 hover:text-gray-700"
                      >
                        <X className="w-3 h-3" />
                      </button>
                    </span>
                  ))}
                </div>
              </div>

              {/* Quiz Behavior Settings */}
              <div className="grid grid-cols-2 gap-4">
                <label className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    checked={exportSettings.allowBackNavigation}
                    onChange={(e) => setExportSettings({
                      ...exportSettings,
                      allowBackNavigation: e.target.checked
                    })}
                    className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                  />
                  <span className="text-sm text-gray-700">Allow back navigation</span>
                </label>

                <label className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    checked={exportSettings.showProgressBar}
                    onChange={(e) => setExportSettings({
                      ...exportSettings,
                      showProgressBar: e.target.checked
                    })}
                    className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                  />
                  <span className="text-sm text-gray-700">Show progress bar</span>
                </label>

                <label className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    checked={exportSettings.randomizeQuestions}
                    onChange={(e) => setExportSettings({
                      ...exportSettings,
                      randomizeQuestions: e.target.checked
                    })}
                    className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                  />
                  <span className="text-sm text-gray-700">Randomize questions</span>
                </label>
              </div>

              {/* Optional Settings */}
              <div className="grid grid-cols-3 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Time Limit (minutes)
                  </label>
                  <input
                    type="number"
                    value={exportSettings.timeLimit || ''}
                    onChange={(e) => setExportSettings({
                      ...exportSettings,
                      timeLimit: e.target.value ? parseInt(e.target.value) : undefined
                    })}
                    placeholder="Optional"
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Max Attempts
                  </label>
                  <input
                    type="number"
                    value={exportSettings.maxAttempts || ''}
                    onChange={(e) => setExportSettings({
                      ...exportSettings,
                      maxAttempts: e.target.value ? parseInt(e.target.value) : undefined
                    })}
                    placeholder="Unlimited"
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Passing Score (%)
                  </label>
                  <input
                    type="number"
                    value={exportSettings.passingScore || ''}
                    onChange={(e) => setExportSettings({
                      ...exportSettings,
                      passingScore: e.target.value ? parseInt(e.target.value) : undefined
                    })}
                    placeholder="Optional"
                    min="0"
                    max="100"
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
              </div>
            </div>

            {/* Export Result */}
            {exportResult && (
              <div className={cn(
                'mt-6 p-4 rounded-lg flex items-start gap-3',
                exportResult.success 
                  ? 'bg-green-50 border border-green-200' 
                  : 'bg-red-50 border border-red-200'
              )}>
                {exportResult.success ? (
                  <CheckCircle className="w-5 h-5 text-green-600 flex-shrink-0 mt-0.5" />
                ) : (
                  <AlertTriangle className="w-5 h-5 text-red-600 flex-shrink-0 mt-0.5" />
                )}
                <div className="flex-1">
                  <p className={cn(
                    'font-medium',
                    exportResult.success ? 'text-green-800' : 'text-red-800'
                  )}>
                    {exportResult.message}
                  </p>
                  {exportResult.errors && exportResult.errors.length > 0 && (
                    <ul className="mt-2 text-sm text-red-700 list-disc list-inside space-y-1">
                      {exportResult.errors.map((error, index) => (
                        <li key={index}>{error}</li>
                      ))}
                    </ul>
                  )}
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Footer */}
        <div className="flex items-center justify-between p-6 border-t border-gray-200 bg-gray-50">
          <button
            onClick={onClose}
            className="px-4 py-2 text-gray-700 hover:bg-gray-200 rounded-lg"
          >
            Cancel
          </button>
          <div className="flex gap-2">
            <button
              onClick={() => handleExport('json')}
              disabled={isExporting || items.length === 0}
              className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <FileJson className="w-4 h-4" />
              {isExporting ? 'Exporting...' : 'Export JSON'}
            </button>
            <button
              onClick={() => handleExport('bundle')}
              disabled={isExporting || items.length === 0}
              className="flex items-center gap-2 px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <Download className="w-4 h-4" />
              {isExporting ? 'Exporting...' : 'Export Bundle'}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};