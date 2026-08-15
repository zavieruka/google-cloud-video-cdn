import { VideoDetail } from "../../../components/video-detail";

export default function VideoPage({ params }: { params: { id: string } }) {
  return <VideoDetail videoId={params.id} />;
}
