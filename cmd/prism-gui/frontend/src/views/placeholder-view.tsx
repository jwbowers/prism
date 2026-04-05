import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/card'
import { Alert, AlertDescription } from '../components/ui/alert'

interface PlaceholderViewProps {
  title: string
  description: string
}

export function PlaceholderView({ title, description }: PlaceholderViewProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="flex flex-col items-center gap-4 py-8 text-center">
          <p className="font-semibold">{title}</p>
          <p className="text-muted-foreground">{description}</p>
          <Alert className="max-w-lg">
            <AlertDescription>This feature will be available in a future update.</AlertDescription>
          </Alert>
        </div>
      </CardContent>
    </Card>
  )
}
