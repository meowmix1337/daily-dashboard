import React, { useState, useEffect } from 'react';
import { useUserSettings, useNewsCategories, useSettingsMutations } from '../hooks/useUserSettings';

interface SettingsModalProps {
  onClose: () => void;
}

export function SettingsModal({ onClose }: SettingsModalProps): React.ReactElement {
  const { settings } = useUserSettings();
  const { data: categoriesData } = useNewsCategories();
  const { save, saveCategories } = useSettingsMutations();

  const [latitude, setLatitude] = useState<string>('');
  const [longitude, setLongitude] = useState<string>('');
  const [timezone, setTimezone] = useState<string>('');
  const [calendarIcsUrl, setCalendarIcsUrl] = useState<string>('');
  const [selectedCategoryIds, setSelectedCategoryIds] = useState<string[]>([]);
  const [focusedField, setFocusedField] = useState<string | null>(null);

  useEffect(() => {
    if (settings) {
      setLatitude(settings.latitude !== null ? String(settings.latitude) : '');
      setLongitude(settings.longitude !== null ? String(settings.longitude) : '');
      setTimezone(settings.timezone ?? '');
      setCalendarIcsUrl(settings.calendar_ics_url ?? '');
    }
  }, [settings]);

  useEffect(() => {
    if (categoriesData) {
      setSelectedCategoryIds(categoriesData.selected.map((c) => c.id));
    }
  }, [categoriesData]);

  function toggleCategory(id: string): void {
    setSelectedCategoryIds((prev) =>
      prev.includes(id) ? prev.filter((c) => c !== id) : [...prev, id]
    );
  }

  function handleSave(): void {
    const body: {
      latitude?: number | null;
      longitude?: number | null;
      timezone?: string | null;
      calendar_ics_url?: string | null;
    } = {
      latitude: latitude !== '' ? parseFloat(latitude) : null,
      longitude: longitude !== '' ? parseFloat(longitude) : null,
      timezone: timezone !== '' ? timezone : null,
      calendar_ics_url: calendarIcsUrl !== '' ? calendarIcsUrl : null,
    };
    save.mutate(body);
    saveCategories.mutate(selectedCategoryIds);
    onClose();
  }

  function handleOverlayClick(e: React.MouseEvent<HTMLDivElement>): void {
    if (e.target === e.currentTarget) {
      onClose();
    }
  }

  const inputStyle = (field: string): React.CSSProperties => ({
    background: 'rgba(255,255,255,0.06)',
    border: `1px solid ${focusedField === field ? 'rgba(99,102,241,0.5)' : 'rgba(255,255,255,0.1)'}`,
    borderRadius: 8,
    padding: '8px 12px',
    fontSize: 13,
    color: 'var(--text-primary)',
    outline: 'none',
    width: '100%',
    boxSizing: 'border-box' as const,
  });

  const sectionLabelStyle: React.CSSProperties = {
    fontSize: 12,
    fontWeight: 600,
    color: 'var(--text-secondary)',
    letterSpacing: '0.08em',
    textTransform: 'uppercase',
    marginBottom: 10,
  };

  const available = categoriesData?.available ?? [];

  return (
    <div
      onClick={handleOverlayClick}
      style={{
        position: 'fixed',
        inset: 0,
        zIndex: 1000,
        background: 'rgba(0,0,0,0.65)',
        backdropFilter: 'blur(6px)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: 24,
      }}
    >
      <div
        style={{
          background: 'rgba(20,20,35,0.92)',
          border: '1px solid rgba(255,255,255,0.12)',
          borderRadius: 20,
          backdropFilter: 'blur(24px)',
          padding: 28,
          width: '100%',
          maxWidth: 520,
          maxHeight: '90vh',
          overflowY: 'auto',
          display: 'flex',
          flexDirection: 'column',
          gap: 24,
        }}
      >
        {/* Header */}
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <span style={{ fontSize: 20, fontWeight: 700, color: 'var(--text-primary)' }}>
            Settings
          </span>
          <button
            onClick={onClose}
            style={{
              fontSize: 20,
              background: 'none',
              border: 'none',
              cursor: 'pointer',
              color: 'var(--text-secondary)',
              lineHeight: 1,
              padding: 4,
            }}
            aria-label="Close settings"
          >
            ✕
          </button>
        </div>

        {/* Section: Location */}
        <div>
          <div style={sectionLabelStyle}>Location</div>
          <div style={{ display: 'flex', gap: 12 }}>
            <div style={{ flex: 1 }}>
              <label style={{ fontSize: 12, color: 'var(--text-muted)', display: 'block', marginBottom: 6 }}>
                Latitude
              </label>
              <input
                type="number"
                step={0.0001}
                min={-90}
                max={90}
                value={latitude}
                onChange={(e) => setLatitude(e.target.value)}
                onFocus={() => setFocusedField('latitude')}
                onBlur={() => setFocusedField(null)}
                style={inputStyle('latitude')}
                placeholder="37.7749"
              />
            </div>
            <div style={{ flex: 1 }}>
              <label style={{ fontSize: 12, color: 'var(--text-muted)', display: 'block', marginBottom: 6 }}>
                Longitude
              </label>
              <input
                type="number"
                step={0.0001}
                min={-180}
                max={180}
                value={longitude}
                onChange={(e) => setLongitude(e.target.value)}
                onFocus={() => setFocusedField('longitude')}
                onBlur={() => setFocusedField(null)}
                style={inputStyle('longitude')}
                placeholder="-122.4194"
              />
            </div>
          </div>
        </div>

        {/* Section: Timezone */}
        <div>
          <div style={sectionLabelStyle}>Timezone</div>
          <input
            type="text"
            value={timezone}
            onChange={(e) => setTimezone(e.target.value)}
            onFocus={() => setFocusedField('timezone')}
            onBlur={() => setFocusedField(null)}
            style={inputStyle('timezone')}
            placeholder="e.g. America/New_York"
          />
        </div>

        {/* Section: Calendar ICS URL */}
        <div>
          <div style={sectionLabelStyle}>Calendar ICS URL</div>
          <input
            type="url"
            value={calendarIcsUrl}
            onChange={(e) => setCalendarIcsUrl(e.target.value)}
            onFocus={() => setFocusedField('calendarIcsUrl')}
            onBlur={() => setFocusedField(null)}
            style={inputStyle('calendarIcsUrl')}
            placeholder="https://..."
          />
          <div style={{ fontSize: 11, color: 'var(--text-muted)', marginTop: 6 }}>
            Your private Google/Outlook calendar feed URL
          </div>
        </div>

        {/* Section: News Categories */}
        <div>
          <div style={sectionLabelStyle}>News Categories</div>
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
            {available.map((category) => {
              const isSelected = selectedCategoryIds.includes(category.id);
              return (
                <button
                  key={category.id}
                  onClick={() => toggleCategory(category.id)}
                  style={{
                    background: isSelected ? 'rgba(99,102,241,0.2)' : 'rgba(255,255,255,0.05)',
                    border: isSelected
                      ? '1px solid rgba(99,102,241,0.5)'
                      : '1px solid rgba(255,255,255,0.1)',
                    borderRadius: 8,
                    padding: '6px 12px',
                    fontSize: 12,
                    cursor: 'pointer',
                    color: isSelected ? 'var(--text-accent)' : 'var(--text-secondary)',
                  }}
                >
                  {category.label}
                </button>
              );
            })}
          </div>
        </div>

        {/* Footer */}
        <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 10, paddingTop: 4 }}>
          <button
            onClick={onClose}
            style={{
              background: 'transparent',
              border: '1px solid rgba(255,255,255,0.15)',
              borderRadius: 8,
              padding: '9px 24px',
              fontSize: 14,
              color: 'var(--text-secondary)',
              cursor: 'pointer',
            }}
          >
            Cancel
          </button>
          <button
            onClick={handleSave}
            style={{
              background: '#6366f1',
              border: 'none',
              borderRadius: 8,
              padding: '9px 24px',
              fontSize: 14,
              fontWeight: 600,
              color: '#fff',
              cursor: 'pointer',
            }}
          >
            Save
          </button>
        </div>
      </div>
    </div>
  );
}
